# Radigo Schedule

**Please do not use this project for commercial use. Only for your personal, non-commercial use.**  
**個人での視聴の目的以外で利用しないでください.**

This tool automatically calculates the most recent past broadcast time for radio programs defined in a `schedule.json` file and directly records them using the `go-radiko` and `radigo` Go libraries.

## Features

- Reads a schedule of radio programs from a `schedule.json` file.
- For each program, calculates the most recent past broadcast time.
- Directly records the program by integrating with `go-radiko` (for API interactions and stream URLs) and `radigo` (for M3U8 chunklist parsing) Go libraries.
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
    ./radigoSchedule
    ```

    You can also specify a different path for the schedule file using the `--file` flag:

    ```bash
    ./radigoSchedule --file /path/to/your/schedule.json
    ```
    Recorded files will be saved in the `output/` directory.

## Schedule File Configuration

### `schedule.json`

This file contains the list of programs you want to record. It's an array of JSON objects, where each object has the following properties:

- `program_name`: The name of the program (for logging purposes).
- `day_of_week`: The day of the week in Japanese ("日", "月", "火", "水", "木", "金", "土").
- `start_time`: The start time of the program in `HHMMSS` format (e.g., "030000" for 3:00 AM).
- `station_id`: The station ID used by Radiko (e.g., "LFR").

**Example `schedule.json`:**

```json
[
  {
    "program_name": "My Favorite Show",
    "day_of_week": "水",
    "start_time": "030000",
    "station_id": "LFR"
  },
  {
    "program_name": "Another Show",
    "day_of_week": "金",
    "start_time": "113000",
    "station_id": "FMT"
  }
]
```


