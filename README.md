# Go Web Crawler

## Overview

This is an enhanced web crawler built in Go, extending upon [@afazio1's original web-crawler project](https://github.com/afazio1/web-crawler). The crawler has been significantly enhanced with robots.txt compliance, headless Chrome browser support for better JavaScript rendering, a real-time web interface for monitoring and searching, and improved architecture for scalability.

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

### 🎯 **Real-time Web Interface**
- Live statistics dashboard with WebSocket updates
- Google-like search interface with relevance scoring
- Real-time crawling progress monitoring
- Responsive design for desktop and mobile

### 🔍 **Advanced Search Engine**
- Intelligent relevance scoring algorithm
- Search across page titles, content, and URLs
- Color-coded score indicators (high/medium/low relevance)
- Pagination support for large result sets

### 🏗️ **Enhanced Architecture**
- Clean, modular design with separate packages for different concerns
- RESTful API server with WebSocket support
- Configurable via environment variables
- MongoDB integration for scalable data storage
- Comprehensive statistics tracking

## Features

### Core Crawling
- **Concurrent Crawling**: Multi-threaded crawling with configurable limits
- **Duplicate Prevention**: URL deduplication using hash-based crawled set
- **Content Extraction**: Extracts page titles and meaningful content
- **Error Handling**: Graceful handling of network errors, timeouts, and invalid URLs

### Web Interface & API
- **Live Statistics Dashboard**: Real-time crawling metrics and progress updates
- **Search Interface**: Google-like search with relevance scoring
- **REST API**: Endpoints for statistics, search, and page retrieval
- **WebSocket Integration**: Real-time updates every 5 seconds
- **Responsive Design**: Works on desktop and mobile devices

### Search & Scoring
- **Relevance Scoring**: Intelligent algorithm weighing title, content, and URL matches
- **Smart Sorting**: Results automatically ranked by relevance
- **Visual Indicators**: Color-coded score badges for easy relevance assessment
- **Full-text Search**: Search across all page content with regex support

### Data Storage
- **MongoDB Storage**: Scalable document storage with search capabilities
- **Configurable**: Environment-based configuration for different deployment scenarios
- **Statistics Tracking**: Real-time crawling statistics and performance metrics

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

### Basic Crawling with Web Interface

Run the crawler with the web interface:

```bash
go run cmd/crawler/main.go
```

The crawler will:
1. Start the web interface on `http://localhost:8080`
2. Begin crawling from the configured `SEED_URL`
3. Check robots.txt compliance before crawling each URL
4. Use headless Chrome to render pages
5. Extract content and discover new URLs
6. Store results in MongoDB with real-time updates

### Accessing the Web Interface

1. **Start the crawler**: `go run cmd/crawler/main.go`
2. **Open your browser**: Navigate to `http://localhost:8080`
3. **Monitor progress**: Watch live statistics update in real-time
4. **Search crawled pages**: Use the search interface to find specific content

### Web Interface Features

#### 📊 Live Statistics Dashboard
- **Pages Crawled**: Total number of pages processed
- **Queue Size**: Current URLs waiting to be crawled
- **Crawl Rate**: Pages processed per minute
- **Uptime**: Time since crawler started

#### 🔍 Search Interface
- **Smart Search**: Search across page titles, content, and URLs
- **Relevance Scoring**: Results ranked by relevance with color-coded badges
- **Pagination**: Navigate through large result sets
- **Real-time Results**: Search updates as new pages are crawled

#### 📋 Recent Pages
- **Live Updates**: Recently crawled pages appear automatically
- **Page Preview**: See page titles, URLs, and content snippets
- **Direct Links**: Click URLs to visit original pages

## API Endpoints

The crawler provides a REST API for programmatic access:

### Statistics
- `GET /api/stats` - Get current crawling statistics
- `WebSocket /ws/stats` - Real-time statistics updates

### Search
- `GET /api/search?q=query&page=1` - Search crawled pages
- `GET /api/pages?page=1&limit=10` - Get recent pages

### Project Structure

```
go-webcrawler/
├── cmd/
│   └── crawler/          # Main application entry point
├── internal/
│   ├── api/             # REST API server and WebSocket handlers
│   ├── config/          # Configuration management
│   ├── crawler/         # Core crawling logic (fetcher & parser)
│   ├── models/          # Data structures (Page with scoring)
│   ├── queue/           # URL queue and crawled set management
│   ├── robots/          # Robots.txt handling
│   ├── stats/           # Statistics tracking
│   ├── storage/         # Database interfaces and MongoDB implementation
│   └── utils/           # Utility functions
├── web/
│   └── static/          # Web interface (HTML, CSS, JavaScript)
├── go.mod
├── go.sum
└── README.md
```

## Key Components

### Crawler (`internal/crawler/`)
- **Fetcher**: Uses headless Chrome via chromedp for page rendering
- **Parser**: Extracts content, titles, and discovers new URLs

### API Server (`internal/api/`)
- **REST Endpoints**: Statistics, search, and page retrieval
- **WebSocket Handler**: Real-time updates for web interface
- **Static File Serving**: Serves web interface files

### Search Engine (`internal/storage/`)
- **MongoDB Integration**: Full-text search with regex support
- **Relevance Scoring**: Intelligent algorithm for result ranking
- **Pagination**: Efficient handling of large result sets

### Web Interface (`web/static/`)
- **Dashboard**: Real-time statistics and monitoring
- **Search Interface**: Google-like search with scoring
- **Responsive Design**: Works on all devices

### Robots Checker (`internal/robots/`)
- Fetches and parses robots.txt files
- Caches robots.txt per domain
- Supports user-agent specific rules and crawl delays

### Queue Management (`internal/queue/`)
- Thread-safe URL queue implementation
- Hash-based duplicate URL detection
- Memory-efficient crawled set tracking

## Search Scoring Algorithm

The search engine uses an intelligent scoring system:

### **Title Matches (Highest Weight)**
- Base score: **10 points**
- Exact match: **+20 points**
- Starts with query: **+5 points**
- Each occurrence: **+3 points**

### **URL Matches (Medium Weight)**
- Base score: **5 points**
- Domain match: **+2 points**
- Each occurrence: **+2 points**

### **Content Matches (Lower Weight)**
- Base score: **1 point**
- Starts with query: **+2 points**
- Each occurrence: **+0.5 points**

### **Visual Indicators**
- 🟢 **Green (15+ points)**: Highly relevant
- 🟡 **Yellow (5-15 points)**: Moderately relevant
- 🔴 **Red (0-5 points)**: Less relevant

## Statistics

The crawler tracks and displays:
- Pages crawled per minute
- Crawl-to-queue ratio
- Total URLs queued vs. crawled
- Real-time progress updates
- Search query performance

## Docker Support

Start MongoDB with Docker:
```bash
docker run -d --name mongodb -p 27017:27017 mongo:latest
```

## Acknowledgments

This project builds upon the excellent foundation laid by [@afazio1](https://github.com/afazio1) in their [original web-crawler project](https://github.com/afazio1/web-crawler). The core crawling concepts and MongoDB integration patterns were inspired by their work.

## License

This project is open source and available under the [MIT License](LICENSE).

## Future Enhancements

- [ ] Distributed crawling capabilities
- [ ] Configurable crawling algorithms (BFS, DFS, priority-based)
- [ ] Advanced search filters and operators
- [ ] Export functionality for search results
- [ ] Crawler scheduling and automation
- [ ] Performance analytics and reporting