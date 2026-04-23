use crate::commands::http_client::HttpClient;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Project {
    pub id: String,
    pub name: String,
    pub path: String,
    pub language: String,
    #[serde(rename = "created_at")]
    pub created_at: String,
    #[serde(rename = "updated_at")]
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateProjectRequest {
    pub name: String,
    pub path: String,
    pub language: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateProjectRequest {
    pub name: Option<String>,
    pub language: Option<String>,
    #[serde(rename = "build_config")]
    pub build_config: Option<serde_json::Value>,
    #[serde(rename = "deploy_config")]
    pub deploy_config: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
pub struct ApiResponse<T> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

#[derive(Deserialize)]
struct ProjectListResponse {
    projects: Vec<Project>,
    total: i32,
}

#[tauri::command]
pub async fn list_projects() -> Result<Vec<Project>, String> {
    let client = HttpClient::new();

    let response: ApiResponse<ProjectListResponse> = client
        .get("/api/v1/projects")
        .await
        .map_err(|e| format!("List projects failed: {}", e))?;

    response.data.map(|r| r.projects).ok_or_else(|| "No data".to_string())
}

#[tauri::command]
pub async fn create_project(request: CreateProjectRequest) -> Result<Project, String> {
    let client = HttpClient::new();

    let response: ApiResponse<Project> = client
        .post("/api/v1/projects", &request)
        .await
        .map_err(|e| format!("Create project failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn get_project(project_id: String) -> Result<Option<Project>, String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/projects/{}", project_id);

    let response: ApiResponse<Project> = client
        .get(&path)
        .await
        .map_err(|e| format!("Get project failed: {}", e))?;

    Ok(response.data)
}

#[tauri::command]
pub async fn update_project(project_id: String, request: UpdateProjectRequest) -> Result<Project, String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/projects/{}", project_id);

    let response: ApiResponse<Project> = client
        .put(&path, &request)
        .await
        .map_err(|e| format!("Update project failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}

#[tauri::command]
pub async fn delete_project(project_id: String) -> Result<(), String> {
    let client = HttpClient::new();
    let path = format!("/api/v1/projects/{}", project_id);

    let _: ApiResponse<serde_json::Value> = client
        .delete(&path)
        .await
        .map_err(|e| format!("Delete project failed: {}", e))?;

    Ok(())
}
