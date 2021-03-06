package endpoints

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/tomochain/tomox-sdk/interfaces"
	"github.com/tomochain/tomox-sdk/types"
	"github.com/tomochain/tomox-sdk/ws"
)

type PriceBoardEndpoint struct {
	priceBoardService interfaces.PriceBoardService
}

// ServePriceBoardResource sets up the routing of token endpoints and the corresponding handlers.
func ServePriceBoardResource(
	r *mux.Router,
	priceBoardService interfaces.PriceBoardService,
) {
	e := &PriceBoardEndpoint{priceBoardService}

	ws.RegisterChannel(ws.PriceBoardChannel, e.handlePriceBoardWebSocket)
}

func (e *PriceBoardEndpoint) handlePriceBoardWebSocket(input interface{}, c *ws.Client) {
	socket := ws.GetPriceBoardSocket()
	errInvalidPayload := map[string]string{"Message": "Invalid payload"}
	if input == nil {
		socket.SendErrorMessage(c, errInvalidPayload)
		return
	}
	b, _ := json.Marshal(input)
	var ev *types.WebsocketEvent

	err := json.Unmarshal(b, &ev)
	if err != nil {
		logger.Error(err)
		return
	}
	if ev == nil {
		socket.SendErrorMessage(c, errInvalidPayload)
		return
	}

	if ev.Type != types.SUBSCRIBE && ev.Type != types.UNSUBSCRIBE {
		logger.Info("Event Type", ev.Type)
		socket.SendErrorMessage(c, errInvalidPayload)
		return
	}

	b, _ = json.Marshal(ev.Payload)
	var p *types.SubscriptionPayload

	err = json.Unmarshal(b, &p)
	if err != nil {
		logger.Error(err)
		msg := map[string]string{"Message": "Internal server error"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if ev.Type == types.SUBSCRIBE {
		if p == nil {
			socket.SendErrorMessage(c, errInvalidPayload)
			return
		}
		if (p.BaseToken == common.Address{}) {
			msg := map[string]string{"Message": "Invalid base token"}
			socket.SendErrorMessage(c, msg)
			return
		}

		if (p.QuoteToken == common.Address{}) {
			msg := map[string]string{"Message": "Invalid quote token"}
			socket.SendErrorMessage(c, msg)
			return
		}

		e.priceBoardService.Subscribe(c, p.BaseToken, p.QuoteToken)
	}

	if ev.Type == types.UNSUBSCRIBE {
		if p == nil {
			e.priceBoardService.Unsubscribe(c)
			return
		}

		e.priceBoardService.UnsubscribeChannel(c, p.BaseToken, p.QuoteToken)
	}
}
