from io import BytesIO
from pathlib import Path
from typing import Optional, Union, List
import asyncio
import aioboto3
from botocore.config import Config
from app.config import settings


class AsyncS3BucketService:
    def __init__(
            self,
            bucket_name: settings.s3.bucket_name,
            endpoint: settings.s3.endpoint,
            access_key: settings.s3.access_key,
            secret_key: settings.s3.secret_key,
    ) -> None:
        self.bucket_name = bucket_name
        self.endpoint = endpoint
        self.access_key = access_key
        self.secret_key = secret_key
        self._session = aioboto3.Session()

    async def create_s3_client(self):
        """Создает асинхронного S3 клиента"""
        client = await self._session.client(
            "s3",
            endpoint_url=self.endpoint,
            aws_access_key_id=self.access_key,
            aws_secret_access_key=self.secret_key,
            config=Config(signature_version="s3v4"),
        ).__aenter__()
        return client

    async def upload_file_object(
            self,
            prefix: str,
            source_file_name: str,
            content: Union[str, bytes],
    ) -> None:
        """Асинхронная загрузка файла в S3"""
        async with await self.create_s3_client() as client:
            destination_path = str(Path(prefix, source_file_name))

            if isinstance(content, bytes):
                buffer = BytesIO(content)
            else:
                buffer = BytesIO(content.encode("utf-8"))

            await client.upload_fileobj(buffer, self.bucket_name, destination_path)

    async def list_objects(self, prefix: str) -> List[str]:
        """Асинхронное получение списка объектов"""
        async with await self.create_s3_client() as client:
            response = await client.list_objects_v2(
                Bucket=self.bucket_name,
                Prefix=prefix
            )

            storage_content: List[str] = []
            try:
                contents = response["Contents"]
            except KeyError:
                return storage_content

            for item in contents:
                storage_content.append(item["Key"])

            return storage_content

    async def delete_file_object(self, prefix: str, source_file_name: str) -> None:
        """Асинхронное удаление файла"""
        async with await self.create_s3_client() as client:
            path_to_file = str(Path(prefix, source_file_name))
            await client.delete_object(Bucket=self.bucket_name, Key=path_to_file)

    async def download_file_object(self, prefix: str, source_file_name: str) -> bytes:
        """Асинхронное скачивание файла"""
        async with await self.create_s3_client() as client:
            path_to_file = str(Path(prefix, source_file_name))
            response = await client.get_object(Bucket=self.bucket_name, Key=path_to_file)

            async with response['Body'] as stream:
                content = await stream.read()
                return content

    async def generate_presigned_url(
            self,
            prefix: str,
            source_file_name: str,
            expiration: int = 3600,
            method: str = "get_object"
    ) -> str:
        """Генерация предварительно подписанного URL"""
        async with await self.create_s3_client() as client:
            path_to_file = str(Path(prefix, source_file_name))

            url = await client.generate_presigned_url(
                ClientMethod=method,
                Params={
                    'Bucket': self.bucket_name,
                    'Key': path_to_file
                },
                ExpiresIn=expiration
            )
            return url

    async def object_exists(self, prefix: str, source_file_name: str) -> bool:
        """Проверка существования объекта"""
        async with await self.create_s3_client() as client:
            path_to_file = str(Path(prefix, source_file_name))
            try:
                await client.head_object(Bucket=self.bucket_name, Key=path_to_file)
                return True
            except Exception:
                return False