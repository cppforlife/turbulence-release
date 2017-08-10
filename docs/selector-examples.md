# Selectors

## Examples

- Select all instances in AZ `z1` within a deployment `cf`:

```json
{
  "AZ": {
    "Name": "z1"
  },
  "Deployment": {
    "Name": "cf"
  }
}
```

- Select random 50% of instances from `postgres` instance group:

```json
{
  "Group": {
    "Name": "postgres"
  },
  "ID": {
    "Limit": "50%"
  }
}
```

- Select one particular instance by ID:

```json
{
  "ID": {
    "Values": ["53c5ae69-4622-4103-9766-230adcf3baef"]
  }
}
```
