package handlers

import "context"

func (h *Handlers) HandleDisconnect(ctx context.Context) {
	if h.cancel != nil && ctx.Err() == nil {
		h.cancel(nil)
	}

	if !h.client.IsEmptyConnection() {
		h.client.Close()
	}
	h.IsConnected = false

	h.updater.UpdateConnectStatus(false)
}
