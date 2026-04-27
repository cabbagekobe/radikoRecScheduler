# radikoRecScheduler

> Powered by [`uru2/rec_radiko_ts`](https://github.com/uru2/rec_radiko_ts) — the actual recording is delegated to that shell script. This tool invokes `rec_radiko_ts.sh` (MIT License) as an external process. `rec_radiko_ts` is not bundled with this project and must be installed separately. See [Acknowledgements](#謝辞--acknowledgements) below.

**This project is intended for personal, non-commercial use only. Commercial use is prohibited.**  
**個人での視聴・利用目的以外での使用は禁止します。**

This tool automatically calculates the most recent past broadcast time for radio programs defined in a `schedule.json` file and records them by delegating to the [`uru2/rec_radiko_ts`](https://github.com/uru2/rec_radiko_ts) shell script.

## Features

- Reads a schedule of radio programs from a `schedule.json` file.
- For each program, calculates the most recent past broadcast time.
- Builds the corresponding radiko timefree URL and invokes `rec_radiko_ts.sh` to record it.
- Saves recordings as `.m4a` files under the `output/` directory.

## Requirements

- **Go**: Version 1.22 or higher to build and run this application.
- **`rec_radiko_ts.sh`**: The shell script from [`uru2/rec_radiko_ts`](https://github.com/uru2/rec_radiko_ts), placed on your `PATH` (or referenced via `RADIKO_REC_TS_SCRIPT`) — see the installation steps below.
- **Runtime tools** required by `rec_radiko_ts.sh`: `ffmpeg` (3.x or later), `curl`, and `xmllint` (from `libxml2`).

## Installing `rec_radiko_ts`

`rec_radiko_ts` is not packaged. Clone it and place `rec_radiko_ts.sh` somewhere on your `PATH` — that is the recommended setup and requires no extra configuration. As an escape hatch, you can instead point the `RADIKO_REC_TS_SCRIPT` environment variable at the script's absolute path.

### macOS (Homebrew)

```bash
# 1. Install runtime tools
brew install ffmpeg libxml2 curl

# 2. Clone rec_radiko_ts somewhere persistent
git clone https://github.com/uru2/rec_radiko_ts ~/.local/share/rec_radiko_ts

# 3. Expose rec_radiko_ts.sh via PATH (recommended)
ln -s ~/.local/share/rec_radiko_ts/rec_radiko_ts.sh /usr/local/bin/rec_radiko_ts.sh
# Alternatively, add the clone directory to PATH in ~/.zshrc:
#   export PATH="$HOME/.local/share/rec_radiko_ts:$PATH"
```

### Debian / Ubuntu

```bash
sudo apt update
sudo apt install -y ffmpeg curl libxml2-utils git

git clone https://github.com/uru2/rec_radiko_ts ~/.local/share/rec_radiko_ts

# Expose rec_radiko_ts.sh via PATH (recommended)
sudo ln -s ~/.local/share/rec_radiko_ts/rec_radiko_ts.sh /usr/local/bin/rec_radiko_ts.sh
# Alternatively, add the clone directory to PATH in ~/.bashrc:
#   export PATH="$HOME/.local/share/rec_radiko_ts:$PATH"
```

### Using `RADIKO_REC_TS_SCRIPT` instead of PATH

If you cannot (or prefer not to) modify `PATH`, set the absolute path of the script via an environment variable:

```bash
export RADIKO_REC_TS_SCRIPT=$HOME/.local/share/rec_radiko_ts/rec_radiko_ts.sh
```

When set, `RADIKO_REC_TS_SCRIPT` takes precedence over `PATH` lookup.

### Updating `rec_radiko_ts`

```bash
git -C ~/.local/share/rec_radiko_ts pull
```

### Verifying your installation

After installing, run the bundled checker — it confirms that `rec_radiko_ts.sh` and every runtime tool can be found:

```bash
./radikoRecScheduler --check
```

Sample output when everything is in place:

```
Checking dependencies for radikoRecScheduler:

  [OK]      rec_radiko_ts.sh   -> /usr/local/bin/rec_radiko_ts.sh
  [OK]      ffmpeg             -> /opt/homebrew/bin/ffmpeg
  [OK]      curl               -> /usr/bin/curl
  [OK]      xmllint            -> /opt/homebrew/opt/libxml2/bin/xmllint

Result: All dependencies OK.
```

If anything is missing, the line is marked `[MISSING]` together with an install hint, and the command exits with status `1` so it can be wired into CI or pre-flight scripts.

## Usage

1.  **Build the application:**

    ```bash
    go build
    ```

2.  **Create your schedule:**

    Create a `schedule.json` file with your desired programs. See the format below.

3.  **Run the application:**

    ```bash
    ./radikoRecScheduler
    ```

    Available flags:

    - `--file <path>`: Use a custom schedule file (defaults to the XDG config path described below).
    - `--check`: Verify that `rec_radiko_ts.sh` and the required runtime tools (`ffmpeg`, `curl`, `xmllint`) are installed, then exit. Returns `0` when everything is found and `1` otherwise.

    Recorded files are saved in the `output/` directory as `{YYYYMMDDHHMMSS}-{stationID}-{programName}.m4a`.

## Schedule File Configuration

### `schedule.json` Location

The `schedule.json` file, which defines your radio program schedule, is searched for in the following order:

1.  **XDG Base Directory (Recommended):**
    *   The application first checks the path specified by the `XDG_CONFIG_HOME` environment variable. If set, it will look for `schedule.json` at `$XDG_CONFIG_HOME/radikoRecScheduler/schedule.json`.
    *   If `XDG_CONFIG_HOME` is not set, it defaults to `~/.config/radikoRecScheduler/schedule.json` on Linux/macOS.
    *   On Windows, this typically resolves to `%APPDATA%\radikoRecScheduler\schedule.json`.
    *   The necessary directory structure (`radikoRecScheduler` within the config directory) will be created automatically if it doesn't exist.

2.  **Current Working Directory (Fallback):**
    *   If `schedule.json` is not found in the XDG Base Directory compliant location, the application will then look for `schedule.json` in the current directory where `radikoRecScheduler` is executed.

3.  **Custom Path (Using `--file` flag):**
    *   You can always specify a custom path to your `schedule.json` using the `--file` flag:
        ```bash
        ./radikoRecScheduler --file /path/to/your/custom/schedule.json
        ```

This file contains the list of programs you want to record. It's an array of JSON objects, where each object has the following properties:

- `program_name`: The name of the program (for logging purposes).
- `day_of_week`: The day of the week in Japanese ("日", "月", "火", "水", "木", "金", "土").
- `start_time`: The start time of the program in `HHMMSS` format (e.g., "030000" for 3:00 AM).
- `station_id`: The station ID used by Radiko (e.g., "LFR").

**Example `schedule.json`:**

```json
[
  {
    "program_name": "オードリーのオールナイトニッポン",
    "day_of_week": "土",
    "start_time": "010000",
    "station_id": "LFR"
  },
  {
    "program_name": "櫻坂46 こちら有楽町星空放送局",
    "day_of_week": "日",
    "start_time": "230000",
    "station_id": "LFR"
  }
]
```

## 謝辞 / Acknowledgements

録音処理は [`uru2/rec_radiko_ts`](https://github.com/uru2/rec_radiko_ts) ([MIT License](https://github.com/uru2/rec_radiko_ts/blob/master/LICENSE.md)) に頼っています（同梱はしていないので別途インストールが必要）。

[@uru2](https://github.com/uru2) さんと `rec_radiko_ts` のコントリビューターのみなさんに感謝します。
