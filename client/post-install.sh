pwd=$(pwd)
if [[ $pwd == *"node_modules"* ]]; then
  echo "Installed as Node module. Building package..."
  npm run build-pkg
  else
    echo "Installed as stand-alone UI. Skipped building package"
fi
