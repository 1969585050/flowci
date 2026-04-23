use serde::Serialize;
use thiserror::Error;

pub type Result<T> = std::result::Result<T, FlowCIError>;

#[derive(Error, Debug)]
pub enum FlowCIError {
    #[error("Project error: {0}")]
    Project(String),

    #[error("Build error: {0}")]
    Build(String),

    #[error("Deploy error: {0}")]
    Deploy(String),

    #[error("Docker error: {0}")]
    Docker(String),

    #[error("Config error: {0}")]
    Config(String),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),

    #[error("Network error: {0}")]
    Network(String),

    #[error("Validation error: {0}")]
    Validation(String),
}

#[derive(Serialize)]
pub struct ErrorResponse {
    pub code: i32,
    pub message: String,
}

impl FlowCIError {
    pub fn error_code(&self) -> i32 {
        match self {
            FlowCIError::Project(_) => 1003,
            FlowCIError::Build(_) => 2002,
            FlowCIError::Deploy(_) => 2003,
            FlowCIError::Docker(_) => 2001,
            FlowCIError::Config(_) => 3003,
            FlowCIError::Io(_) => 3001,
            FlowCIError::Json(_) => 3001,
            FlowCIError::Network(_) => 4003,
            FlowCIError::Validation(_) => 1001,
        }
    }

    pub fn to_response(&self) -> ErrorResponse {
        ErrorResponse {
            code: self.error_code(),
            message: self.to_string(),
        }
    }
}

impl From<reqwest::Error> for FlowCIError {
    fn from(err: reqwest::Error) -> Self {
        FlowCIError::Network(err.to_string())
    }
}
