# Installing RepoMind

**Language:** English | [简体中文](INSTALL.zh-CN.md)

RepoMind is CLI-first. The recommended install path depends on whether you want to use source builds or release binaries.

## Run From Source

Use this while developing RepoMind:

```powershell
go run ./cmd/repomind analyze .
```

## Build Locally

Build a binary for the current platform:

```powershell
go build -o repomind ./cmd/repomind
```

On Windows:

```powershell
.\repomind.exe version
.\repomind.exe analyze .
```

On macOS or Linux:

```bash
./repomind version
./repomind analyze .
```

## Install With Go

After the repository is published, install from the module path:

```bash
go install github.com/patrick892368/RepoMind/cmd/repomind@latest
```

Make sure your Go binary directory is on `PATH`.

Windows PowerShell:

```powershell
$env:PATH += ";$(go env GOPATH)\bin"
repomind version
```

macOS or Linux:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
repomind version
```

## Install From Release Binary

Download the archive for your OS and CPU from GitHub Releases.

Archives contain:

```txt
repomind
README.md
.env.example
```

Windows:

1. Extract the `.zip`.
2. Move `repomind.exe` to a directory such as `C:\Tools\repomind`.
3. Add that directory to the user `PATH`.
4. Open a new terminal and run:

```powershell
repomind version
repomind analyze .
```

macOS or Linux:

```bash
tar -xzf repomind-<version>-<os>-<arch>.tar.gz
chmod +x repomind
sudo mv repomind /usr/local/bin/repomind
repomind version
repomind analyze .
```

## AI Provider Keys

RepoMind runs offline by default. Network AI providers are optional.

Create a local `.env` in the repository you analyze:

```env
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GEMINI_API_KEY=
GROK_API_KEY=
HTTPS_PROXY=http://127.0.0.1:10809
```

`.env` is ignored by Git. Do not commit API keys.
