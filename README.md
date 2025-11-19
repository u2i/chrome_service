# Chrome Service

A REST API service that provides PDF generation and screenshot capabilities using headless Chrome.

## Endpoints

### POST /v1/print

Convert HTML to PDF.

**Request:**
```json
{
  "html": "<html><body>Hello World</body></html>",
  "options": {
    "landscape": false,
    "print_background": true,
    "scale": 1.0,
    "paper_width": 8.5,
    "paper_height": 11,
    "margin_top": 0.4,
    "margin_bottom": 0.4,
    "margin_left": 0.4,
    "margin_right": 0.4
  }
}
```

**Response:** PDF file (application/pdf)

### POST /v1/print_pdfa

Convert HTML to PDF/A.

**Request:** Same as `/v1/print`

**Response:** PDF/A file (application/pdf)

### POST /v1/convert_pdfa

Convert an existing PDF to PDF/A.

**Request:**
```json
{
  "pdf": "base64-encoded-pdf-content",
  "options": {
    "pdfa_version": "3b"
  }
}
```

**Response:** PDF/A file (application/pdf)

### POST /v1/screenshot

Capture a screenshot of HTML content.

**Request:**
```json
{
  "html": "<html><body>Hello World</body></html>",
  "options": {
    "width": 1280,
    "height": 720,
    "full_page": false
  }
}
```

**Response:** PNG image (image/png)

## Running

### Development

```bash
go run main.go
```

### Docker

```bash
docker build -t chrome-service .
docker run -p 8080:8080 chrome-service
```

## Environment Variables

- `PORT`: HTTP server port (default: 8080)
