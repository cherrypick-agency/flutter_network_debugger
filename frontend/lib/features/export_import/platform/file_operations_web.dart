// Web-specific implementation that re-exports dart:html
// This file is only used on web platform where dart:html is available.

// Re-export all dart:html types used in export/import feature
export 'dart:html'
    show
        File,
        FileReader,
        FileUploadInputElement,
        DataTransfer,
        Event,
        Document,
        Window,
        document,
        window;
