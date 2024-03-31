# AVSC to JSON Converter
This script is a simple utility written in Go that converts Avro schema files (.avsc) to JSON format.

### Usage
To use this script, you need to pass the .avsc file as a command-line argument:

Replace <input.avsc> with the path to your Avro schema file.

### How It Works
The script reads the Avro schema file specified as a command-line argument. It then converts the schema to JSON format and writes the output to a new file with the same name as the input file but with a .json extension.

For example, if you run go run main.go example.avsc, the script will create a new file named example.json with the JSON representation of the Avro schema.

### Error Handling
If the script encounters an error while reading the Avro schema file or converting it to JSON, it will print an error message and exit.

### Contributing
Contributions are welcome. Please submit a pull request if you have any improvements or bug fixes.