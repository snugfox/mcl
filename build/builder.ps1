Set-StrictMode -Version 3.0
$InformationPreference = "Continue"

function Get-StringHash {
	[OutputType([string])]
	param (
		[Parameter(Mandatory)]
		[string]
		$Value
	)

	$TempFile = New-TemporaryFile
	Set-Content $TempFile $Value -NoNewline
	$Hash = $(Get-FileHash $TempFile -Algorithm SHA256).Hash
	$TempFile.Delete()
	return $Hash
}

$Dockerfile = "./build/builder/Dockerfile"
$Name       = "snugfox/mcl-builder"
$Tag        = $(Get-StringHash `
								$(-Join $(Get-Content ./go.sum,$Dockerfile -Raw)) `
							). `
							Substring(0, 7). `
							ToLowerInvariant()
$Image      = "${Name}:${Tag}"

$InitScriptBlock = {
	docker image inspect -f "{{ .Id }}" "$Image" *>&1 | Out-Null
	if (! $?) {
		docker pull -q "$Image" *>&1 | Out-Null
		if ($?) {
			Write-Information "Pulled ${Image}"
		} else {
			docker build -q -f "$Dockerfile" -t "$Image" . | Out-Null
			Write-Information "Built ${Image}"
		}
	}
}

$Command = ($args.Length -gt 0) ? $args[0] : ""
switch ($Command) {
	"image" {
		Write-Host $Image
	}
	"push" {
		Write-Warning "It is advised not to push builder images built on Windows. To force a push, use the push-force command."
	}
	"push-force" {
		& $InitScriptBlock
		docker push "${Image}" | Out-Null
		Write-Information "Pushed ${Image}"
	}
	"run" {
		& $InitScriptBlock
		docker run --rm `
			-e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD `
			-e GITHUB_TOKEN `
			-v //var/run/docker.sock:/var/run/docker.sock `
			-v "${PWD}:/go/src/github.com/snugfox/mcl" `
			"$Image" $args[1..($args.Length-1)]
	}
}
