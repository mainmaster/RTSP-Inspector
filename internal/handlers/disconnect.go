package handlers

func (h *Handlers) disconnect() {
	if h.cancel != nil {
		h.cancel()
	}

	if !h.client.IsEmptyConnection() {
		h.client.Close()
	}
	h.isConnected = false

	h.ui.UpdateConnectStatus(false)
}
