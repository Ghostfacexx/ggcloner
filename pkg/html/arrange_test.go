package html

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/goclone-dev/goclone/pkg/crawler"
	"github.com/goclone-dev/goclone/pkg/file"
	"github.com/goclone-dev/goclone/testutils"
)

// TestArrange verifies that the LinkRestructure function correctly reorganizes the paths
// of resources (CSS, JS and images) in the HTML file, ensuring that:
// 1. Paths are correctly updated to their new locations
// 2. Files exist in their expected locations
// 3. Original element attributes are preserved
func TestArrange(t *testing.T) {
	testutils.SilenceStdoutInTests()
	ts := testutils.NewArrangeTestServer()
	defer ts.Close()

	// Initial setup
	projectDirectory := file.CreateProject("test")
	defer os.RemoveAll(projectDirectory)

	// Run crawler and restructuring
	crawler.Collector(context.Background(), ts.URL, projectDirectory, nil, "", "")

	if err := LinkRestructure(projectDirectory); err != nil {
		t.Fatalf("Error during restructuring: %v", err)
	}

	// Verify that index.html exists
	if !file.Exists(projectDirectory + "/index.html") {
		t.Fatal("index.html file should exist")
	}

	// Get and verify file content
	indexFileContent := file.GetFileContent(projectDirectory + "/index.html")
	if indexFileContent == testutils.ArrangeIndexContent {
		t.Fatalf("Expected restructured HTML, not original: %s", testutils.ArrangeIndexContent)
	}

	// Verify that files exist in expected locations
	expectedFiles := []string{
		"/css/index.css",
		"/js/index.js",
		"/imgs/image.png",
	}

	for _, expectedFile := range expectedFiles {
		if !file.Exists(projectDirectory + expectedFile) {
			t.Fatalf("File %s should exist", expectedFile)
		}
	}

	// Verify paths in HTML
	expectedPaths := []string{
		"css/index.css",
		"js/index.js",
		"imgs/image.png",
	}

	for _, path := range expectedPaths {
		if !strings.Contains(indexFileContent, path) {
			t.Fatalf("Expected to find path %s in HTML", path)
		}
	}

	// Verify that original attributes are preserved
	if !strings.Contains(indexFileContent, `alt="Red dot"`) {
		t.Fatal("Expected to preserve alt attribute in image")
	}
}

// TestArrangeMultipleHTMLFiles verifies that the LinkRestructure function correctly
// processes multiple HTML files in the project directory
func TestArrangeMultipleHTMLFiles(t *testing.T) {
	testutils.SilenceStdoutInTests()

	// Initial setup
	projectDirectory := file.CreateProject("test_multiple")
	defer os.RemoveAll(projectDirectory)

	// Create multiple HTML files with resource links
	htmlFiles := map[string]string{
		"index.html": `<html>
<link rel="stylesheet" href="style.css">
<script src="app.js"></script>
<img src="logo.png" alt="Logo" />
</html>`,
		"about.html": `<html>
<link rel="stylesheet" href="about-style.css">
<script src="about-app.js"></script>
<img src="about-image.jpg" alt="About" />
</html>`,
		"contact.html": `<html>
<link rel="stylesheet" href="contact.css">
<script src="contact.js"></script>
<img src="contact.gif" alt="Contact" />
</html>`,
	}

	// Write HTML files
	for filename, content := range htmlFiles {
		if err := os.WriteFile(projectDirectory+"/"+filename, []byte(content), 0777); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// Run restructuring
	if err := LinkRestructure(projectDirectory); err != nil {
		t.Fatalf("Error during restructuring: %v", err)
	}

	// Verify each HTML file has been restructured
	expectedRestructuredPaths := map[string][]string{
		"index.html": {
			"css/style.css",
			"js/app.js",
			"imgs/logo.png",
		},
		"about.html": {
			"css/about-style.css",
			"js/about-app.js",
			"imgs/about-image.jpg",
		},
		"contact.html": {
			"css/contact.css",
			"js/contact.js",
			"imgs/contact.gif",
		},
	}

	for filename, expectedPaths := range expectedRestructuredPaths {
		content := file.GetFileContent(projectDirectory + "/" + filename)
		for _, path := range expectedPaths {
			if !strings.Contains(content, path) {
				t.Fatalf("Expected to find path %s in %s", path, filename)
			}
		}
	}

	// Verify that original attributes are preserved
	indexContent := file.GetFileContent(projectDirectory + "/index.html")
	if !strings.Contains(indexContent, `alt="Logo"`) {
		t.Fatal("Expected to preserve alt attribute in index.html")
	}

	aboutContent := file.GetFileContent(projectDirectory + "/about.html")
	if !strings.Contains(aboutContent, `alt="About"`) {
		t.Fatal("Expected to preserve alt attribute in about.html")
	}

	contactContent := file.GetFileContent(projectDirectory + "/contact.html")
	if !strings.Contains(contactContent, `alt="Contact"`) {
		t.Fatal("Expected to preserve alt attribute in contact.html")
	}
}
