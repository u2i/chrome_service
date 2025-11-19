package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type PrintRequest struct {
	HTML    string                 `json:"html" binding:"required"`
	Options map[string]interface{} `json:"options"`
}

type ConvertRequest struct {
	PDF     string                 `json:"pdf" binding:"required"` // base64 encoded PDF
	Options map[string]interface{} `json:"options"`
}

type ScreenshotRequest struct {
	HTML    string                 `json:"html" binding:"required"`
	Options map[string]interface{} `json:"options"`
}

func main() {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.POST("/print", handlePrint)
		v1.POST("/print_pdfa", handlePrintPDFA)
		v1.POST("/convert_pdfa", handleConvertPDFA)
		v1.POST("/screenshot", handleScreenshot)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Chrome service on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func handlePrint(c *gin.Context) {
	var req PrintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdf, err := htmlToPDF(req.HTML, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdf)
}

func handlePrintPDFA(c *gin.Context) {
	var req PrintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdf, err := htmlToPDF(req.HTML, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to PDF/A using ghostscript
	pdfa, err := convertToPDFA(pdf, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdfa)
}

func handleConvertPDFA(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Decode base64 PDF
	pdfBytes := []byte(req.PDF)

	// Convert to PDF/A using ghostscript
	pdfa, err := convertToPDFA(pdfBytes, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdfa)
}

func handleScreenshot(c *gin.Context) {
	var req ScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	screenshot, err := htmlToScreenshot(req.HTML, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", screenshot)
}

func htmlToPDF(html string, options map[string]interface{}) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuffer []byte

	// Build print options
	printParams := page.PrintToPDF()

	if options != nil {
		if landscape, ok := options["landscape"].(bool); ok {
			printParams = printParams.WithLandscape(landscape)
		}
		if displayHeaderFooter, ok := options["display_header_footer"].(bool); ok {
			printParams = printParams.WithDisplayHeaderFooter(displayHeaderFooter)
		}
		if printBackground, ok := options["print_background"].(bool); ok {
			printParams = printParams.WithPrintBackground(printBackground)
		}
		if scale, ok := options["scale"].(float64); ok {
			printParams = printParams.WithScale(scale)
		}
		if paperWidth, ok := options["paper_width"].(float64); ok {
			printParams = printParams.WithPaperWidth(paperWidth)
		}
		if paperHeight, ok := options["paper_height"].(float64); ok {
			printParams = printParams.WithPaperHeight(paperHeight)
		}
		if marginTop, ok := options["margin_top"].(float64); ok {
			printParams = printParams.WithMarginTop(marginTop)
		}
		if marginBottom, ok := options["margin_bottom"].(float64); ok {
			printParams = printParams.WithMarginBottom(marginBottom)
		}
		if marginLeft, ok := options["margin_left"].(float64); ok {
			printParams = printParams.WithMarginLeft(marginLeft)
		}
		if marginRight, ok := options["margin_right"].(float64); ok {
			printParams = printParams.WithMarginRight(marginRight)
		}
		if pageRanges, ok := options["page_ranges"].(string); ok {
			printParams = printParams.WithPageRanges(pageRanges)
		}
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = printParams.Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBuffer, nil
}

func htmlToScreenshot(html string, options map[string]interface{}) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var screenshotBuffer []byte

	// Default viewport
	width := int64(1280)
	height := int64(720)
	fullPage := false

	if options != nil {
		if w, ok := options["width"].(float64); ok {
			width = int64(w)
		}
		if h, ok := options["height"].(float64); ok {
			height = int64(h)
		}
		if fp, ok := options["full_page"].(bool); ok {
			fullPage = fp
		}
	}

	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(width, height),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
	}

	if fullPage {
		tasks = append(tasks, chromedp.FullScreenshot(&screenshotBuffer, 100))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&screenshotBuffer))
	}

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return screenshotBuffer, nil
}

func convertToPDFA(pdfBytes []byte, options map[string]interface{}) ([]byte, error) {
	// This is a placeholder - you'll need to implement ghostscript conversion
	// For now, we'll just return the original PDF
	// In production, you'd call ghostscript to convert to PDF/A

	// Example ghostscript command for PDF/A conversion:
	// gs -dPDFA=3 -dBATCH -dNOPAUSE -sColorConversionStrategy=RGB \
	//    -sDEVICE=pdfwrite -dPDFACompatibilityPolicy=1 \
	//    -sOutputFile=output.pdf input.pdf

	log.Println("Warning: PDF/A conversion not implemented yet, returning original PDF")
	return pdfBytes, nil
}
