from dependency_client import etcd_client

watch_events,cancel = etcd_client.get("aaa")
print(watch_events)