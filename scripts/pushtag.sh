# check if $1 is stage or prod
if [ "$1" != "stage" ] && [ "$1" != "prod" ]; then
  echo "Invalid parameter supplied. Please provide 'stage' or 'prod' as argument."
  exit 1
fi

git fetch --tags
latest_tag=$(git tag -l | grep "$1" | sort -V | tail -n 1 | cut -d' ' -f1)

while true; do
  new_tag=$(echo $latest_tag | awk -F. '{$NF+=1} 1' OFS=.)
  if ! git rev-parse "$new_tag" >/dev/null 2>&1; then
    break
  else
    latest_tag=$new_tag
  fi
done

echo "New tag: $new_tag"
git push
git tag $new_tag
git push --tags
#git push origin $new_tag
