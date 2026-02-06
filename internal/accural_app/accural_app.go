package accuralapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Oleg2210/gophermart/internal/domain"
	"github.com/Oleg2210/gophermart/internal/serializers"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	clientTimeout     = 10
	noOrdersTimeSleep = 0
	workersCount      = 3
	clientMaxRetries  = 3
)

type Sleeper struct {
	mu    sync.Mutex
	until time.Time
}

func (s *Sleeper) SleepUntil(d time.Duration) {
	s.mu.Lock()
	t := time.Now().Add(d)
	if t.After(s.until) {
		s.until = t
	}
	s.mu.Unlock()
}

func (s *Sleeper) Wait(ctx context.Context) error {
	s.mu.Lock()
	until := s.until
	s.mu.Unlock()

	if time.Now().After(until) {
		return nil
	}

	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type Accrual struct {
	ctx     context.Context
	baseURL string
	service domain.Service
	logger  zap.Logger
	sleeper Sleeper
	job     chan string
}

func NewAccrual(ctx context.Context, baseURL string, service domain.Service, logger zap.Logger) *Accrual {

	return &Accrual{
		ctx:     ctx,
		baseURL: baseURL,
		service: service,
		logger:  logger,
		sleeper: Sleeper{until: time.Now()},
		job:     make(chan string, workersCount),
	}
}

func StartAccural(ctx context.Context, baseURL string, service domain.Service, logger zap.Logger) {
	select {
	case <-ctx.Done():
		logger.Info("closing accural jobs creator within context singal")
		return
	default:
	}

	a := NewAccrual(ctx, baseURL, service, logger)
	defer close(a.job)

	for i := 0; i < workersCount; i++ {
		go processJob(a)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("closing accural jobs creator within context singal")
			return
		default:
		}

		orders, err := service.GetUnprocessedOrders(ctx, workersCount)

		if err != nil {
			logger.Error("failed to get unprocessed orders", zap.Error(err))
			continue
		}

		if len(orders) == 0 {
			time.Sleep(time.Duration(noOrdersTimeSleep) * time.Second)
		}

		for _, o := range orders {
			a.job <- o.ID
		}
	}
}

func processJob(a *Accrual) {
	client := NewClient(a.baseURL)

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("closing accural worker within context singal")
			return
		case orderID := <-a.job:
			processOrder(a, orderID, *client)
		}
	}
}

func processOrder(a *Accrual, orderID string, client Client) {
	for i := 0; i < clientMaxRetries; i++ {
		select {
		case <-a.ctx.Done():
			a.logger.Info("closing accural worker within context singal")
			return
		default:
		}

		err := a.sleeper.Wait(a.ctx)
		if err != nil {
			return
		}

		orderResp, sleepDuration, err := client.GetOrder(a.ctx, orderID)
		if err != nil {
			if errors.Is(err, ErrRepeatable) {
				a.logger.Error("failed to atempt to get order: ", zap.Error(err))
				continue
			}
			a.logger.Error("failed to process order", zap.Error(err))
			return
		}

		if orderResp != nil {
			amount := decimal.NewFromInt(0)

			if orderResp.Accrual != nil {
				amount = *orderResp.Accrual
			}

			err = a.service.ProcessAccural(a.ctx, orderID, orderResp.Status, amount)

			if err != nil {
				a.logger.Error("failed to process accural", zap.Error(err))
			}
			return
		}

		if sleepDuration > 0 {
			a.sleeper.SleepUntil(time.Duration(sleepDuration) * time.Second)
		}
	}
}

var ErrRepeatable = errors.New("repeatable error")

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: clientTimeout * time.Second,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*serializers.AccrualResponse, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {

	case http.StatusOK:
		var order serializers.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return nil, 0, err
		}
		return &order, 0, nil

	case http.StatusNoContent:
		return nil, 0, nil

	case http.StatusNotFound:
		return nil, 0, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		sec, err := strconv.Atoi(retryAfter)
		if err != nil {
			return nil, 0, err
		}
		return nil, sec, nil

	default:
		return nil, 0, errors.Join(fmt.Errorf("accrual unexpected status: %d", resp.StatusCode), ErrRepeatable)
	}
}
