use crate::commands::http_client::HttpClient;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct PushRequest {
    #[serde(rename = "image_tag")]
    pub image_tag: String,
}

#[derive(Debug, Deserialize)]
pub struct ApiResponse<T> {
    pub code: i32,
    pub message: String,
    pub data: Option<T>,
}

#[derive(Debug, Deserialize)]
pub struct PushResponse {
    pub image_tag: String,
    pub status: String,
}

#[tauri::command]
pub async fn push_image(request: PushRequest) -> Result<PushResponse, String> {
    let client = HttpClient::new();

    let response: ApiResponse<PushResponse> = client
        .post("/api/v1/push", &request)
        .await
        .map_err(|e| format!("Push image failed: {}", e))?;

    response.data.ok_or_else(|| "No data in response".to_string())
}
