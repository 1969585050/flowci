use crate::modules::builder::{Builder, BuildConfig, BuildResult};
use crate::error::Result;
use serde::{Deserialize, Serialize};
use chrono::Utc;
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct BuildRequest {
    pub project_id: String,
    pub language: String,
    pub image_tag: String,
    pub registry_address: String,
    pub registry_namespace: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct BuildResponse {
    pub id: String,
    pub image_id: String,
    pub tag: String,
    pub status: String,
    pub logs: Vec<String>,
    pub started_at: String,
    pub finished_at: String,
}

#[tauri::command]
pub async fn build_image(request: BuildRequest) -> Result<BuildResponse> {
    let config = BuildConfig {
        project_id: request.project_id.clone(),
        language: serde_json::from_str(&format!("\"{}\"", request.language))
            .unwrap_or(crate::modules::builder::ProgrammingLanguage::Custom),
        build_source: crate::modules::builder::BuildSource {
            source_type: crate::modules::builder::BuildSourceType::Local,
            path: None,
            git_url: None,
            git_branch: None,
            git_credentials_id: None,
        },
        base_image: None,
        build_command: None,
        environment_vars: Default::default(),
        tags: vec![request.image_tag.clone()],
        registry: crate::modules::builder::RegistryConfig {
            registry_type: crate::modules::builder::RegistryType::AliyunAcr,
            address: request.registry_address.clone(),
            namespace: request.registry_namespace.clone(),
            credentials_id: None,
        },
    };

    let dockerfile = Builder::generate_dockerfile(&config);
    let context = Builder::prepare_build_context(&config.build_source, &dockerfile)?;

    let result = Builder::build(&config, context).await?;

    Ok(BuildResponse {
        id: result.id,
        image_id: result.image_id,
        tag: result.tag,
        status: format!("{:?}", result.status).to_lowercase(),
        logs: vec![],
        started_at: result.started_at.to_rfc3339(),
        finished_at: result.finished_at.to_rfc3339(),
    })
}

#[tauri::command]
pub async fn get_build_logs(build_id: String) -> Result<Vec<String>> {
    Ok(vec![format!("Build {} logs placeholder", build_id)])
}

#[tauri::command]
pub async fn list_builds(project_id: String) -> Result<Vec<BuildResponse>> {
    Ok(vec![])
}
