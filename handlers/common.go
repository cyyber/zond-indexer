package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"syscall"

	"github.com/Prajjawalk/zond-indexer/price"
	"github.com/Prajjawalk/zond-indexer/services"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetCurrency(r *http.Request) string {
	if cookie, err := r.Cookie("currency"); err == nil {
		return cookie.Value
	}

	return "ETH"
}

func GetCurrentPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return price.GetEthRoundPrice(price.GetEthPrice("USD"))
	}

	if cookie.Value == "ETH" {
		return price.GetEthRoundPrice(price.GetEthPrice("USD"))
	}
	return price.GetEthRoundPrice(price.GetEthPrice(cookie.Value))
}

func GetCurrentPriceFormatted(r *http.Request) template.HTML {
	userAgent := r.Header.Get("User-Agent")
	userAgent = strings.ToLower(userAgent)
	price := GetCurrentPrice(r)
	if strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "windows phone") {
		return utils.KFormatterEthPrice(price)
	}
	return utils.FormatAddCommas(uint64(price))
}

func GetCurrentPriceKFormatted(r *http.Request) template.HTML {
	return utils.KFormatterEthPrice(GetCurrentPrice(r))
}

func GetCurrencySymbol(r *http.Request) string {

	cookie, err := r.Cookie("currency")
	if err != nil {
		return "$"
	}

	switch cookie.Value {
	case "AUD":
		return "A$"
	case "CAD":
		return "C$"
	case "CNY":
		return "¥"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "¥"
	case "RUB":
		return "₽"
	default:
		return "$"
	}
}

// used to handle errors constructed by Template.ExecuteTemplate correctly
func handleTemplateError(w http.ResponseWriter, r *http.Request, fileIdentifier string, functionIdentifier string, infoIdentifier string, err error) error {
	// ignore network related errors
	if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, syscall.ETIMEDOUT) {
		logger.WithFields(logrus.Fields{
			"file":       fileIdentifier,
			"function":   functionIdentifier,
			"info":       infoIdentifier,
			"error type": fmt.Sprintf("%T", err),
			"route":      r.URL.String(),
		}).WithError(err).Error("error executing template")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	}
	return err
}

// LatestState will return common information that about the current state of the eth2 chain
func LatestState(c *gin.Context) {
	w := c.Writer
	r := c.Request
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", utils.Config.Chain.Config.SecondsPerSlot)) // set local cache to the seconds per slot interval
	currency := GetCurrency(r)
	data := services.LatestState()
	// data.Currency = currency
	data.EthPrice = price.GetEthPrice(currency)
	data.EthRoundPrice = price.GetEthRoundPrice(data.EthPrice)
	data.EthTruncPrice = utils.KFormatterEthPrice(data.EthRoundPrice)

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
