package handlers

import "context"

func (h *Handlers) HandelDisconnect(ctx context.Context) {
	if h.cancel != nil {
		h.cancel()
	}

	if !h.client.IsEmptyConnection() {
		h.client.Close()
	}
	h.IsConnected = false

	h.ui.UpdateConnectStatus(false)
}
