from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import BaseModel
from pydantic import PostgresDsn


class S3Settings(BaseModel):
    bucket_name: str = "achive"
    endpoint: str
    access_key: str
    secret_key: str


class RunSettings(BaseModel):
    host: str = "0.0.0.0"
    port: int = 50051
    version: str = "1.0.0"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file="app/.env",
        case_sensitive=False,
        env_nested_delimiter="__",
    )
    run: RunSettings = RunSettings()
    s3: S3Settings


settings = Settings()
