Set-StrictMode -Version 3.0

Write-Host "Building Docker image..."
$imageID = $(docker build -q -f ./build/Dockerfile.build .)
Write-Host "Built image $imageID"

Write-Host "Running make in build image..."
docker run --rm `
	-e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD `
	-e GITHUB_TOKEN `
	-v //var/run/docker.sock:/var/run/docker.sock `
	-v "${PWD}:/go/src/github.com/snugfox/mcl" `
	$imageID make $args
