use crate::modules::docker::{DockerClient, RegistryAuth};
use crate::modules::builder::RegistryType;
use crate::error::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct PushRequest {
    pub image_tag: String,
    pub registry_type: String,
    pub registry_address: String,
    pub registry_namespace: String,
    pub username: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PushResponse {
    pub success: bool,
    pub message: String,
}

#[tauri::command]
pub async fn push_image(request: PushRequest) -> Result<PushResponse> {
    let docker = DockerClient::new().await?;

    let auth = match request.registry_type.as_str() {
        "aliyun" | "acr" => {
            RegistryAuth::new_acr(&request.registry_namespace, &request.username, &request.password)
        }
        "dockerhub" | "docker_hub" => {
            RegistryAuth::new_docker_hub(&request.username, &request.password)
        }
        _ => {
            RegistryAuth {
                server: request.registry_address.clone(),
                username: request.username.clone(),
                password: request.password.clone(),
                email: None,
            }
        }
    };

    let full_tag = format!("{}/{}/{}",
        request.registry_address,
        request.registry_namespace,
        request.image_tag
    );

    docker.push_image(&full_tag, &auth).await?;

    Ok(PushResponse {
        success: true,
        message: format!("Image {} pushed successfully", full_tag),
    })
}
