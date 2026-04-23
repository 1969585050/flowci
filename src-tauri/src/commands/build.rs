use crate::commands::http_client::HttpClient;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct BuildRequest {
    pub project_id: String,
    pub language: String,
    pub context_path: String,
    pub dockerfile_path: Option<String>,
    pub image_tags: Vec<String>,
    #[serde(default)]
    pub build_args: std::collections::HashMap<String, String>,
    #[serde(default)]
    pub no_cache: bool,
    #[serde(default)]
    pub pull_base_image: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct BuildResponse {
    pub id: String,
    pub image_id: String,
    #[serde(rename = "image_id")]
    pub tags: Vec<String>,
    pub size: i64,
    #[serde(rename = "duration_ms")]
    pub duration_ms: i64,
    pub status: String,
    pub logs: Option<Vec<String>>,
    #[serde(rename = "started_at")]
    pub started_at: String,
    #[serde(rename = "finished_at")]
    pub finished_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

#[tauri::command]
pub async fn build_image(request: BuildRequest) -> Result<BuildResponse, String> {
    let client = HttpClient::new();

    let response: ApiResponse<BuildResponse> = client
        .post("/api/v1/builds", &request)
        .await
        .map_err(|e| format!("Build request failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn get_build_logs(build_id: String) -> Result<Vec<String>, String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/builds/{}/logs", build_id);

    #[derive(Deserialize)]
    struct LogsResponse {
        #[serde(rename = "build_id")]
        build_id: String,
        logs: Vec<String>,
    }

    let response: ApiResponse<LogsResponse> = client
        .get(&path)
        .await
        .map_err(|e| format!("Get build logs failed: {}", e))?;

    response.data.map(|r| r.logs).ok_or_else(|| "No data".to_string())
}

#[tauri::command]
pub async fn list_builds() -> Result<Vec<BuildResponse>, String> {
    let client = HttpClient::new();

    #[derive(Deserialize)]
    struct BuildListResponse {
        builds: Vec<BuildResponse>,
        total: i32,
    }

    let response: ApiResponse<BuildListResponse> = client
        .get("/api/v1/builds")
        .await
        .map_err(|e| format!("List builds failed: {}", e))?;

    response.data.map(|r| r.builds).ok_or_else(|| "No data".to_string())
}
