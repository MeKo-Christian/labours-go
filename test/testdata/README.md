# Test Data Files

This directory contains sample hercules protobuf files for testing labours-go.

## Files

- **simple_burndown.pb**: Small-scale test data with sample project data
- **realistic_burndown.pb**: Large-scale test data with more comprehensive metrics
  
## Generated Data Characteristics

### Simple Burndown Data
- Basic project burndown matrix
- Simple file and people data
- Metadata with timestamps

### Realistic Burndown Data  
- More comprehensive burndown analysis
- Multiple developers and files
- Extended time range

## Usage in Tests

These files are used by:
- Unit tests for reader functionality
- Integration tests for end-to-end workflows
- Visual regression tests for chart consistency
- Performance benchmarks

## Regeneration

To regenerate this test data, run:
```bash
go run test/create_sample_data.go
```
