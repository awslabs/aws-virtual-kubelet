
## Metrics
Virtual Kubelet uses Prometheus to collect metrics, Prometheus is an open-source system that collects and stores its metrics as time series data,

----

### Metrics captured within virtual kubelet

"vkec2_pods_created_total"  
"vkec2_ec2_launched_total"  
"vkec2_pods_deleted_total"  
"vkec2_grpc_connection_errors_total"  
"vkec2_grpc_connection_timeout_errors_total"  
"vkec2_ec2_launch_errors_total"  
"vkec2_ec2_termination_errors_total"  
"vkec2_ec2_terminated_total"
"vkec2_create_eni_errors_total"  
"vkec2_delete_eni_errors_total"  
"vkec2_describe_eni_errors_total"  
"vkec2_describe_ec2_errors_total"  
"vkec2_launch_application_grpc_errors_total"  
"vkec2_terminate_application_grpc_errors_total"  
"vkec2_get_application_health_grpc_errors_total"  
"vkec2_get_nodename_errors_total"  
"vkec2_get_pod_from_local_cache_errors_total"  
"vkec2_check_pod_health_grpc_errors_total"  
"vkec2_check_pod_health_nil_response_total"  
"vkec2_pods_deleted_from_health_checks_total"  
"vkec2_get_agent_identity_grpc_errors_total"  
"vkec2_create_ca_cert_errors_total"  
"vkec2_create_cert_signed_by_ca_errors_total"  
"vkec2_retrieve_secret_from_k8s_total"  
"vkec2_create_secret_errors_total"  
"vkec2_update_secret_errors_total"  
"vkec2_create_secret_from_k8s_total"  
"vkec2_update_secret_from_k8s_total"  
"vkec2_ca_certs_missing_in_cache_total"  
"vkec2_client_certs_missing_in_cache_total"  
"vkec2_ca_certs_missing_in_secret_total"  
"vkec2_expired_ca_certs_total"
"vkec2_expired_client_certs_total"  
"vkec2_health_checks_reset_pod_status"  
"vkec2_watch_application_health_stream_errors_total"  
"vkec2_watch_application_health_errors_total"  
"vkec2_check_application_health_errors_total"  
"vkec2_grpc_client_errors_total"  
"vkec2_warm_ec2_launch_errors_total"  
"vkec2_warm_ec2_terminated_total"  
"vkec2_warm_ec2_launched_total"  
"vkec2_warm_ec2_termination_errors_total"  
"vkec2_ec2_create_tag_errors_total"  
"vkec2_health_checks_unhealthy_pod_status"  

### exposed endpoints
* /metrics
* /healthz

### Checking metrics
* run `curl http://{vk-ip}:10256/metrics` from an EC2 instance that is managed by VK
