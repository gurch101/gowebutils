# File Management

[godoc](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/fsutils)

The `fsutils` package provides utilities for managing files in an S3 bucket, including uploading, downloading, and deleting objects.

### Initialization

The file service is automatically initialized when the App is created, provided the following environment variables are set:

```sh
export AWS_S3_BUCKET_NAME="my-bucket"
export AWS_S3_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="my-access-key"
export AWS_SECRET_ACCESS_KEY="my-secret-key"
```

### Usage

#### Uploading a File

```go
func (c *MyController) UploadFile(w http.ResponseWriter, r *http.Request) {
  // Parse the multipart form with a maximum file size of 10 MB.
  r.ParseMultipartForm(10 << 20)

  // Retrieve the file from the form data.
  file, handler, err := r.FormFile("file")
  if err != nil {
    fmt.Println("Error retrieving the file:", err)
    return
  }
  defer file.Close()

  // Upload the file to the S3 bucket.
  location, err := c.app.FileService.UploadFile(handler.Filename, file)
  if err != nil {
    fmt.Println("Error uploading file:", err)
    return
  }

  slog.Info("File uploaded successfully", "location", location)
}
```

#### Downloading a File

```go
func (c *MyController) DownloadFile(w http.ResponseWriter, r *http.Request) {
  // Download the file from the S3 bucket.
  contents, err := c.app.FileService.DownloadFile("1.pdf")
  if err != nil {
    fmt.Println("Error downloading file:", err)
    return
  }

  // Set headers for file download.
  w.Header().Set("Content-Disposition", "attachment; filename=1.pdf")
  w.Header().Set("Content-Type", "application/pdf")

  // Write the file contents to the response.
  _, err = w.Write(contents)
  if err != nil {
    fmt.Println("Error writing file to response:", err)
    return
  }
}
```

#### Deleting a File

```go
func (c *MyController) DeleteFile(w http.ResponseWriter, r *http.Request) {
  // Delete the file from the S3 bucket.
  err := c.app.FileService.DeleteFile("1.pdf")
  if err != nil {
    fmt.Println("Error deleting file:", err)
    return
  }

  // Respond with a success status.
  w.WriteHeader(http.StatusOK)
  slog.Info("File deleted successfully")
}
```
