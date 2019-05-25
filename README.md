# huaweicloud-prometheus-discovery

#Install
```
git clone https://github.com/huaweicloud/huaweicloud-prometheus-discovery
go build
```

#Usage


#Help


#example of file
```
[
 {
  "targets": [
   "10.0.0.1:9100"
  ],
  "labels": {
   "name": "demo152"
  }
 },
 {
  "targets": [
   "10.0.0.2:9100"
  ],
  "labels": {
   "name": "ECS_TEST"
  }
 }
]

```

#example prometheus setting
```
scrape_configs:
- job_name: ecs
  file_sd_configs:
    - files:
      - /path/to/ecs_file_sd.yml
      refresh_interval: 10m
```
