package contextprovider

import (
	"net/http"
	"strconv"
	"time"
)

// TimeProvider annotates the request with whether it is during business hours.
type TimeProvider struct{}

// GetContext returns a flag indicating if the current time is within business hours (9am-5pm).
func (TimeProvider) GetContext(req *http.Request) (map[string]string, error) {
	now := time.Now()
	hour := now.Hour()
	inBusiness := hour >= 9 && hour < 17
	return map[string]string{"business_hours": strconv.FormatBool(inBusiness)}, nil
}
