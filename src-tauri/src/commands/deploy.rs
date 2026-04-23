use crate::commands::http_client::HttpClient;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployRequest {
    #[serde(rename = "project_id")]
    pub project_id: String,
    #[serde(rename = "deployment_type")]
    pub deployment_type: String,
    #[serde(rename = "image_tag")]
    pub image_tag: String,
    #[serde(rename = "container_name")]
    pub container_name: String,
    pub ports: Vec<PortMapping>,
    #[serde(rename = "env_vars")]
    pub env_vars: std::collections::HashMap<String, String>,
    pub volumes: Vec<String>,
    #[serde(rename = "restart_policy")]
    pub restart_policy: String,
    pub replicas: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PortMapping {
    #[serde(rename = "host_port")]
    pub host_port: i32,
    #[serde(rename = "container_port")]
    pub container_port: i32,
    pub protocol: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployResponse {
    pub id: String,
    #[serde(rename = "container_id")]
    pub container_id: String,
    pub name: String,
    pub status: String,
    pub ports: Vec<PortMapping>,
    #[serde(rename = "created_at")]
    pub created_at: String,
}

#[derive(Debug, Deserialize)]
pub struct ApiResponse<T> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

#[tauri::command]
pub async fn deploy_compose(request: DeployRequest) -> Result<DeployResponse, String> {
    let client = HttpClient::new();

    let response: ApiResponse<DeployResponse> = client
        .post("/api/v1/deploys", &request)
        .await
        .map_err(|e| format!("Deploy request failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn get_deploy_status(deploy_id: String) -> Result<DeployResponse, String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/deploys/{}", deploy_id);

    let response: ApiResponse<DeployResponse> = client
        .get(&path)
        .await
        .map_err(|e| format!("Get deploy status failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn rollback_deploy(deploy_id: String) -> Result<(), String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/deploys/{}/rollback", deploy_id);

    #[derive(Deserialize)]
    struct RollbackResponse {
        id: String,
        status: String,
    }

    let _: ApiResponse<RollbackResponse> = client
        .post(&path, &serde_json::json!({}))
        .await
        .map_err(|e| format!("Rollback deploy failed: {}", e))?;

    Ok(())
}
