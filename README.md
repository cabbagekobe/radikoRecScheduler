# Radigo Schedule

This tool automatically calculates the most recent past broadcast time for radio programs defined in a `schedule.json` file and executes a recording command via `radigo`.

## Features

- Reads a schedule of radio programs from a `schedule.json` file.
- For each program, calculates the most recent past broadcast time.
- Executes the `radigo rec` command for the calculated broadcast time.
- The path to the `radigo` executable is configurable.

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

## Configuration

### `schedule.json`

This file contains the list of programs you want to record. It's an array of JSON objects, where each object has the following properties:

- `program_name`: The name of the program (for logging purposes).
- `day_of_week`: The day of the week in Japanese ("日", "月", "火", "水", "木", "金", "土").
- `start_time`: The start time of the program in `HHMMSS` format (e.g., "030000" for 3:00 AM).
- `station_id`: The station ID used by `radigo` (e.g., "LFR").

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

### `config.json`

This file configures the `radigoSchedule` application itself. If it doesn't exist, a default one will be created on the first run.

- `radigo_command_path`: The path to the `radigo` executable. The default is `"radigo"`, which assumes it's in your system's PATH.

**Example `config.json`:**

```json
{
  "radigo_command_path": "/usr/local/bin/radigo"
}
```
