# Standalone Scripts

This directory contains standalone Go scripts that can be run independently of the main application.

## Usage

Each script in this directory is a self-contained Go program that can be run using the `go run` command.

### Example

To run the example script:

```bash
# Basic usage
go run scripts/example.go

# With command line arguments
go run scripts/example.go -name "Alice" -debug
```

## Script Structure

Each script should:
1. Be in the `main` package
2. Have a `main()` function
3. Use proper error handling and logging
4. Include command-line flag support when appropriate
5. Be documented with comments explaining its purpose and usage

## Adding New Scripts

1. Create a new `.go` file in this directory
2. Follow the example script structure
3. Add appropriate documentation
4. Test the script using `go run`

## Best Practices

- Use the `flag` package for command-line arguments
- Implement proper error handling
- Include logging for debugging
- Keep scripts focused on a single task
- Document usage with comments 