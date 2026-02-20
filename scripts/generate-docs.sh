#!/bin/bash
set -e

echo "📚 Activating dartdoc_vitepress from git (latest)..."
dart pub global activate --source git https://github.com/777genius/dartdoc_vitepress.git --git-ref main

echo "📝 Generating docs for workspace..."
cd dart_packages
dart pub global run dartdoc_vitepress \
  --format vitepress \
  --workspace-docs \
  --output ../docs-site

echo "✅ Docs generated successfully!"
echo "📍 Location: docs-site/"
echo "🚀 To preview: cd docs-site && npm run dev"
