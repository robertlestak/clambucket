.env-sample:
	find . -name "*.go" | xargs -I{} grep -o 'os.Getenv("[^"]*")' {} | awk -F'"' '{ print $2"=" }' | sort | uniq > .env-sample