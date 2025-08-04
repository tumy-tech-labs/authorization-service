from authorization import Client

client = Client('http://localhost:8080')
result = client.check_access('default', 'alice', 'file1', 'read')
print(result)
