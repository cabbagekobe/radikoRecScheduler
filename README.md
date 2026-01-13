# radikoRecScheduler

**This project is intended for personal, non-commercial use only. Commercial use is prohibited.**
**個人での視聴・利用目的以外での使用は禁止します。**

This tool automatically calculates the most recent past broadcast time for radio programs defined in a `schedule.json` file and directly records them using the `go-radiko` Go library.

## Features

- Reads a schedule of radio programs from a `schedule.json` file.
- For each program, calculates the most recent past broadcast time.
- Directly records the program by integrating with `go-radiko` (for API interactions, stream URLs, and M3U8 chunklist parsing) Go library.
- Downloads and concatenates AAC audio chunks into a single output file.

## Requirements

- **Go**: Version 1.22 or higher is recommended to build and run the application.

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

    You can also specify a different path for the schedule file using the `--file` flag:

    ```bash
    ./radikoRecScheduler --file /path/to/your/schedule.json
    ```
    Recorded files will be saved in the `output/` directory.

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


