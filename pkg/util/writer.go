package util

import (
	"bytes"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type fileWriterSource interface {
	Stream() (io.ReadCloser, error)
}

type resourceWriterReadCloser struct {
	buffer *bytes.Buffer
}

func (r *resourceWriterReadCloser) Read(p []byte) (n int, err error) {
	return r.buffer.Read(p)
}

func (r *resourceWriterReadCloser) Close() error {
	return nil
}

type resourceWriterSource struct {
	obj     runtime.Object
	printer printers.ResourcePrinter
}

func (r *resourceWriterSource) Stream() (io.ReadCloser, error) {
	buf := bytes.NewBuffer(nil)
	if err := r.printer.PrintObj(r.obj, buf); err != nil {
		return nil, err
	}

	return &resourceWriterReadCloser{buffer: buf}, nil
}

type simpleFileWriter struct {
	printer printers.ResourcePrinter
}

func (f *simpleFileWriter) Write(filepath string, src fileWriterSource) error {
	dest, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer dest.Close()

	readCloser, err := src.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(dest, readCloser)
	return err
}

type MultiSourceFileWriter struct {
	printer printers.ResourcePrinter
}

func (f *MultiSourceFileWriter) WriteFromSource(filepath string, source fileWriterSource) error {
	writer := &simpleFileWriter{printer: f.printer}
	return writer.Write(filepath, source)
}

func (f *MultiSourceFileWriter) WriteFromResource(filepath string, obj runtime.Object) error {
	source := &resourceWriterSource{
		obj:     obj,
		printer: f.printer,
	}

	writer := &simpleFileWriter{printer: f.printer}
	return writer.Write(filepath, source)
}

func NewMultiSourceWriter(printer printers.ResourcePrinter) *MultiSourceFileWriter {
	return &MultiSourceFileWriter{printer: printer}
}
