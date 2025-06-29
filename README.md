# Go Web Crawler

## Overview

This is an enhanced web crawler built in Go, extending upon [@afazio1's original web-crawler project](https://github.com/afazio1/web-crawler). The crawler has been significantly enhanced with robots.txt compliance, headless Chrome browser support for better JavaScript rendering, and improved architecture for scalability.

## Key Enhancements

### 🤖 **Robots.txt Compliance**
- Automatically fetches and respects robots.txt directives
- Supports user-agent specific rules and crawl delays
- Caches robots.txt files for efficient checking

### 🌐 **Headless Chrome Integration**
- Uses Chrome DevTools Protocol via `chromedp` for page rendering
- Better support for Single Page Applications (SPAs)
- More consistent page loading and JavaScript execution
- Handles dynamic content that traditional HTTP fetchers miss

### 🏗️ **Enhanced Architecture**
- Clean, modular design with separate packages for different concerns
- Configurable via environment variables
- MongoDB integration for scalable data storage
- Comprehensive statistics tracking

## Features

- **Concurrent Crawling**: Multi-threaded crawling with configurable limits
- **Duplicate Prevention**: URL deduplication using hash-based crawled set
- **Content Extraction**: Extracts page titles and meaningful content
- **Statistics Tracking**: Real-time crawling statistics and performance metrics
- **MongoDB Storage**: Scalable document storage with search capabilities
- **Configurable**: Environment-based configuration for different deployment scenarios
- **Error Handling**: Graceful handling of network errors, timeouts, and invalid URLs

## Prerequisites

- Go 1.24.4 or later
- MongoDB instance (local or cloud)
- Chrome/Chromium browser (for headless operation)

## Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd go-webcrawler
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables by creating a `.env` file:
```bash
cp .env.example .env
```

## Configuration

Create a `.env` file in the root directory with the following variables:

```env
# MongoDB connection string
MONGO_URI=mongodb://localhost:27017

# Starting URL for crawling
SEED_URL=https://example.com

# User agent string for requests and robots.txt checking
USER_AGENT=YourCrawlerBot/1.0
```

## Usage

### Basic Crawling

Run the crawler with default settings:

```bash
go run cmd/crawler/main.go
```

The crawler will:
1. Start from the configured `SEED_URL`
2. Check robots.txt compliance before crawling each URL
3. Use headless Chrome to render pages
4. Extract content and discover new URLs
5. Store results in MongoDB
6. Continue until it reaches 5,000 pages or runs out of URLs

### Project Structure

```
go-webcrawler/
├── cmd/
│   └── crawler/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── crawler/         # Core crawling logic (fetcher & parser)
│   ├── models/          # Data structures
│   ├── queue/           # URL queue and crawled set management
│   ├── robots/          # Robots.txt handling
│   ├── stats/           # Statistics tracking
│   ├── storage/         # Database interfaces and implementations
│   └── utils/           # Utility functions
├── go.mod
├── go.sum
└── README.md
```

## Key Components

### Crawler (`internal/crawler/`)
- **Fetcher**: Uses headless Chrome via chromedp for page rendering
- **Parser**: Extracts content, titles, and discovers new URLs

### Robots Checker (`internal/robots/`)
- Fetches and parses robots.txt files
- Caches robots.txt per domain
- Supports user-agent specific rules and crawl delays

### Storage (`internal/storage/`)
- MongoDB integration for scalable document storage
- Interface-based design for easy storage backend switching

### Queue Management (`internal/queue/`)
- Thread-safe URL queue implementation
- Hash-based duplicate URL detection
- Memory-efficient crawled set tracking

## Statistics

The crawler tracks and displays:
- Pages crawled per minute
- Crawl-to-queue ratio
- Total URLs queued vs. crawled
- Real-time progress updates

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Acknowledgments

This project builds upon the excellent foundation laid by [@afazio1](https://github.com/afazio1) in their [original web-crawler project](https://github.com/afazio1/web-crawler). The core crawling concepts and MongoDB integration patterns were inspired by their work.

## License

This project is open source and available under the [MIT License](LICENSE).

## Currently Planned Future Enhancements

- [ ] Distributed crawling capabilities
- [ ] Configurable crawling algorithms (BFS, DFS, priority-based)
- [ ] Web interface for monitoring, control, and testing