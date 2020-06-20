param (
		[Parameter(Mandatory)]
		[ValidateSet("image", "init", "push", "push-force", "run")]
		[string]
		$Command = "init",

		[Parameter(ValueFromRemainingArguments)]
		[string[]]
		$RunArgs
)

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

$PushScriptBlock = {
	docker push "${Image}" | Out-Null
	Write-Information "Pushed ${Image}"
}

switch ($Command) {
	"image" {
		Write-Host $Image
	}
	"init" {
		& $InitScriptBlock
	}
	"push" {
		if ($IsWindows) {
			Write-Warning "It is advised not to push builder images built on Windows. To force a push, use the push-force command."
		} else {
			& $InitScriptBlock
			& $PushScriptBlock
		}
	}
	"push-force" {
		& $InitScriptBlock
		& $PushScriptBlock
	}
	"run" {
		& $InitScriptBlock
		docker run --rm `
			-e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD `
			-e GITHUB_TOKEN `
			-v //var/run/docker.sock:/var/run/docker.sock `
			-v "${PWD}:/go/src/github.com/snugfox/mcl" `
			"$Image" $RunArgs
	}
}
