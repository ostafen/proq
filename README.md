# proq

**proq** is an open-source command-line tool that acts as an embedded in-memory Prometheus server. It allows you to pull metrics from any given endpoint that exposes Prometheus metrics and visualize them as plots and histograms directly in the terminal, powered by [termui](https://github.com/gizak/termui). This is especially useful for debugging inside a server or container where setting up a full Prometheus instance is impractical. With `proq`, you can quickly inspect metrics without any external dependencies.

## Features
- 🚀 Pulls and stores Prometheus metrics in memory
- 📊 Displays real-time plots and histograms in the terminal
- ⚡ Lightweight and fast
- 🎛️ Simple CLI interface for querying and visualization

## Installation

### Using Go
```sh
go install github.com/ostafen/proq@latest
```

### Manual Build
```sh
git clone https://github.com/ostafen/proq.git
cd proq
make build
```

## Usage

### Start pulling metrics

```sh
proq http://localhost:9090/metrics
```

## Configuration
You can pass the following flags:
- 🌍 `--endpoint` – The URL of the server exposing Prometheus metrics
- 🔄 `--refresh` – Refresh rate for fetching new metrics (default: 5s)
- 🎨 `--theme` – Customize terminal colors for the UI

## Contributing
Contributions are welcome! To contribute:
1. 🍴 Fork the repository
2. 🌱 Create a feature branch (`git checkout -b feature-name`)
3. 💾 Commit your changes (`git commit -m 'Add new feature'`)
4. 📤 Push to the branch (`git push origin feature-name`)
5. 🔥 Open a Pull Request

## License
📜 This project is licensed under the MIT License.

## Acknowledgments
- 💖 [Prometheus](https://prometheus.io/)
- 🖥️ [termui](https://github.com/gizak/termui)

