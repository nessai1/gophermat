package zip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (writer gzipWriter) Write(bytes []byte) (int, error) {
	return writer.Writer.Write(bytes)
}

func GetZipMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
				logger.Debug("Client doesn't accept gzip format.")
				next.ServeHTTP(writer, request)
				return
			}

			gz, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
			if err != nil {
				logger.Fatal("Gzip encoding level doesn't work!", zap.Error(err))
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Internal error while encode content to gzip: " + err.Error()))
				return
			}

			defer gz.Close()

			writer = gzipWriter{
				ResponseWriter: writer,
				Writer:         gz,
			}
			writer.Header().Set("Content-Encoding", "gzip")

			if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
				request.Body, err = gzip.NewReader(request.Body)
				if err != nil {
					logger.Fatal("Internal error while encode body content to gzip", zap.Error(err))
				}
			}

			next.ServeHTTP(writer, request)
		})
	}
}
