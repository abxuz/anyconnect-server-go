package server

import (
	"mime"
)

func init() {
	mime.AddExtensionType(".sh", "application/x-shellscript")
	mime.AddExtensionType(".txt", "text/plain")
}
