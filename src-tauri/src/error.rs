use thiserror::Error;

#[derive(Error, Debug)]
pub enum FlowCIError {
    #[error("Docker error: {0}")]
    Docker(String),

    #[error("Project error: {0}")]
    Project(String),

    #[error("Build error: {0}")]
    Build(String),

    #[error("Deploy error: {0}")]
    Deploy(String),

    #[error("Config error: {0}")]
    Config(String),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Database error: {0}")]
    Database(String),

    #[error("Serialization error: {0}")]
    Serialization(String),
}

impl serde::Serialize for FlowCIError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(self.to_string().as_ref())
    }
}

pub type Result<T> = std::result::Result<T, FlowCIError>;
