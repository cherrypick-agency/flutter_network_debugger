# Hex Viewer Example

This example demonstrates the basic usage of the `hex_viewer` package.

## Features Demonstrated

- **Basic hex viewer display** with configurable bytes per line and grouping
- **Interactive selection** with visual feedback
- **Dynamic configuration** using sliders to adjust display parameters
- **Theme integration** with both light and dark mode support
- **Real-world data** displaying a JSON object encoded as UTF-8 bytes

## Running the Example

```bash
cd example
flutter run
```

## What's Included

The example shows:
- How to create sample binary data (UTF-8 encoded JSON)
- How to configure the `HexViewer` widget with `HexConfig`
- How to handle selection changes with `onSelectionChanged`
- How to integrate with Material Design theming
- How to create an interactive UI with configuration controls

## Code Highlights

### Creating a HexViewer

```dart
HexViewer(
  data: sampleData,
  config: HexConfig(
    bytesPerLine: 16,
    groupSize: 8,
    colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
  ),
  onSelectionChanged: (selection) {
    print('Selected ${selection?.length ?? 0} bytes');
  },
)
```

### Handling Selection

```dart
ByteSelection? _selection;

// In onSelectionChanged callback:
setState(() {
  _selection = selection;
});

// Display selection info:
if (_selection != null) {
  Text('Selected: ${_selection!.length} bytes');
}
```

## Learn More

See the [hex_viewer documentation](https://pub.dev/packages/hex_viewer) for more details on available configuration options and API usage.
