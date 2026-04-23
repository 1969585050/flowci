use crate::commands::http_client::HttpClient;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DockerCheckResponse {
    pub connected: bool,
    pub version: String,
    #[serde(rename = "api_version")]
    pub api_version: String,
    pub os: String,
    pub arch: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ImageInfo {
    pub id: String,
    pub tags: Vec<String>,
    pub size: i64,
    pub created: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ContainerInfo {
    pub id: String,
    pub names: Vec<String>,
    pub image: String,
    pub state: String,
    pub status: String,
    pub ports: Vec<PortInfo>,
    pub created: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PortInfo {
    #[serde(rename = "host_port")]
    pub host_port: i32,
    #[serde(rename = "container_port")]
    pub container_port: i32,
    pub protocol: String,
}

#[derive(Debug, Deserialize)]
pub struct ApiResponse<T> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

#[derive(Deserialize)]
struct ImageListResponse {
    images: Vec<ImageInfo>,
    total: i32,
}

#[derive(Deserialize)]
struct ContainerListResponse {
    containers: Vec<ContainerInfo>,
    total: i32,
}

#[tauri::command]
pub async fn check_docker() -> Result<DockerCheckResponse, String> {
    let client = HttpClient::new();

    let response: ApiResponse<DockerCheckResponse> = client
        .get("/api/v1/docker/check")
        .await
        .map_err(|e| format!("Docker check failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn list_images() -> Result<Vec<ImageInfo>, String> {
    let client = HttpClient::new();

    let response: ApiResponse<ImageListResponse> = client
        .get("/api/v1/docker/images")
        .await
        .map_err(|e| format!("List images failed: {}", e))?;

    response.data.map(|r| r.images).ok_or_else(|| "No data".to_string())
}

#[tauri::command]
pub async fn list_containers() -> Result<Vec<ContainerInfo>, String> {
    let client = HttpClient::new();

    let response: ApiResponse<ContainerListResponse> = client
        .get("/api/v1/docker/containers")
        .await
        .map_err(|e| format!("List containers failed: {}", e))?;

    response.data.map(|r| r.containers).ok_or_else(|| "No data".to_string())
}
