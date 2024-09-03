# Testing the End-to-End Setup

## Start ELK Stack and Redis
```
docker-compose up -d
```
## Perform CRUD Operations
Use cURL, Postman, or any API client to perform CRUD operations.

1. Create an Item
```
curl -X POST http://localhost:3000/items \
     -H 'Content-Type: application/json' \
     -d '{"name": "Item 1", "description": "First item"}'
```
2. Read All Items
```
curl http://localhost:3000/items
```
3. Update an Item
```
curl -X PUT http://localhost:3000/items/1 \
     -H 'Content-Type: application/json' \
     -d '{"name": "Updated Item 1", "description": "Updated description"}'
```
4. Delete an Item
```
curl -X DELETE http://localhost:3000/items/1
```
## Verify Logs in Kibana
Navigate to Kibana:

Open your browser and go to http://localhost:5601.

### Discover Logs:

1- Click on Discover in the sidebar.
2- Select the app-logs-* index pattern you created earlier.
3- You should see log entries corresponding to your CRUD operations.

#### Sample Log Entry:
Each log entry will contain fields like:

- action: The CRUD action performed.
- item: The item data involved in the action.
- timestamp: When the action occurred.

#### Example Log Entry Structure
```
{
  "@timestamp": "2024-09-04T12:34:56.789Z",
  "action": "create",
  "item": {
    "id": 1,
    "name": "Item 1",
    "description": "First item"
  },
  "timestamp": "2024-09-04T12:34:56.789Z"
}
```

#### Kibana Search
You can use Kibana's search and filtering capabilities to query logs based on these fields. For example:

- Find all create actions:
```
{
  "query": {
    "match": {
      "action": "create"
    }
  }
}
```

- Find logs for a specific item ID:

```{
  "query": {
    "match": {
      "item.id": 1
    }
  }
}
```