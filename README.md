# 📡 RTSP-Inspector

**RTSP-Inspector** is a lightweight, GUI-driven tool designed for deep analysis of RTSP streams. It allows you to monitor client-server interactions in real-time, providing visibility into traffic structure and video data statistics.

---

## ✨ Key Features

*   **Request Flow Monitoring**: Full logging of RTSP methods (`OPTIONS`, `DESCRIBE`, `SETUP`, `PLAY`) and server responses in a human-readable format.
*   **RTP Statistics**: Real-time counter of incoming RTP packets, grouped by payload types.
*   **NALU Analysis**: Detailed statistics for Network Abstraction Layer Units (NALU) for H.264/H.265 (SPS, PPS, IDR, etc.).
*   **Intuitive UI**: A clean, interactive interface built with Fyne, featuring a side-by-side view for inspecting detailed packet contents or SDP descriptions.
*   **Visual Logs**: Clear color-coded indicators for incoming (`▶ RECV`) and outgoing (`◀ SENT`) messages.

---

## 🚀 Getting Started

### Prerequisites
To run the Fyne-based GUI, ensure you have the necessary system dependencies (C compiler and development libraries for your OS) installed. See the [Fyne Setup Guide](https://developer.fyne.io).

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/mainmaster/rtsp-inspector.git
   cd rtsp-inspector
   go build ./cmd/ui
   ```
