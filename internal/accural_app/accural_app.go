package accuralapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Oleg2210/gophermart/internal/domain"
	"github.com/Oleg2210/gophermart/internal/serializers"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	ClientTimeout     = 10
	NoOrdersTimeSleep = 0
)

func StartAccural(ctx context.Context, baseURL string, service domain.Service, logger zap.Logger) {

	go func() {
		client := NewClient(baseURL)

		for {
			select {
			case <-ctx.Done():
				logger.Info("closing accural worker within context singal")
				return
			default:
			}

			orders, err := service.GetUnprocessedOrders(ctx)

			if err != nil {
				logger.Error("failed to get unprocessed orders", zap.Error(err))
				continue
			}

			if len(orders) == 0 {
				time.Sleep(time.Duration(NoOrdersTimeSleep) * time.Second)
			}

			for _, o := range orders {
				orderResp, err := client.GetOrder(ctx, o.ID)

				if err != nil {
					logger.Error("failed to process order", zap.Error(err))
					continue
				}

				if orderResp != nil {
					amount := decimal.NewFromInt(0)

					if orderResp.Accrual != nil {
						amount = *orderResp.Accrual
					}

					err = service.ProcessAccural(ctx, o.ID, orderResp.Status, amount)

					if err != nil {
						logger.Error("failed to process accural", zap.Error(err))
					}
				}
			}
		}
	}()
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: ClientTimeout * time.Second,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*serializers.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {

	case http.StatusOK:
		var order serializers.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return nil, err
		}
		return &order, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusNotFound:
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		sec, err := strconv.Atoi(retryAfter)
		if err != nil {
			return nil, err
		}
		time.Sleep(time.Duration(sec) * time.Second)
		return nil, nil

	default:
		return nil, fmt.Errorf("accrual unexpected status: %d", resp.StatusCode)
	}
}
