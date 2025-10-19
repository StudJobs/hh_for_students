import asyncio
import grpc
from grpc import aio
import aioboto3
from datetime import datetime, timedelta
from achivment_service.app import types_pb2
from achivment_service.app import types_pb2_grpc
from achivment_service.app import service_pb2
from achivment_service.app import service_pb2_grpc
from achivment_service.app import common_pb2
from achivment_service.app import common_pb2_grpc
from achivment_service.app.minio.s3 import AsyncS3BucketService


class GrpcService(service_pb2_grpc.AchievementServiceServicer):
    def __init__(self, s3_service: AsyncS3BucketService):
        self.s3_service = s3_service
        self.s3_client = self.s3_service.create_s3_client()

    async def GetAllAchievements(self, request, context):
        achievement_list = types_pb2.AchievementList()
        try:
            for achievement in self.achievements_meta[request.user_uuid].values():
                achievement_list.achievements.append(achievement)
        except Exception as e:
            await context.abort(grpc.StatusCode.INTERNAL, f"Error: {str(e)}")

        return achievement_list

    async def GetAchievementDownloadUrl(self, request, context):
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        s3_key = f"achievements/{user_uuid}/{achievement_name}"

        try:
            s3_client = await self.ensure_s3_client()

            # Генерируем presigned URL для скачивания
            download_url = await s3_client.generate_presigned_url(
                'get_object',
                Params={
                    'Bucket': self.bucket_name,
                    'Key': s3_key
                },
                ExpiresIn=3600  # URL действителен 1 час
            )

            achievement_url = service_pb2.AchievementUrl(
                url=download_url,
                method="GET",
                headers={"Content-Type": "application/octet-stream"}
            )

            print(f"✅ Download URL generated: {download_url[:50]}...")
            return achievement_url
        except Exception as e:
            print(f"❌ Error generating download URL: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"S3 error: {str(e)}")

    async def GetAchievementUploadUrl(self, request, context):
        """Асинхронно получить signed URL для загрузки в S3"""
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        file_name = request.file_name
        file_type = request.file_type

        s3_key = f"achievements/{user_uuid}/{achievement_name}"

        print(f"📤 Generating upload URL for: {s3_key}")

        try:
            s3_client = await self.ensure_s3_client()

            # Генерируем presigned URL для загрузки
            upload_url = await s3_client.generate_presigned_url(
                'put_object',
                Params={
                    'Bucket': self.bucket_name,
                    'Key': s3_key,
                    'ContentType': file_type,
                },
                ExpiresIn=3600  # URL действителен 1 час
            )

            achievement_url = service_pb2.AchievementUrl(
                url=upload_url,
                method="PUT",
                headers={"Content-Type": file_type}
            )

            print(f"✅ Upload URL generated: {upload_url[:50]}...")
            return achievement_url

        except Exception as e:
            print(f"❌ Error generating upload URL: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"S3 error: {str(e)}")

    async def AddAchievementMeta(self, request, context):
        """Асинхронно сохранить метаданные достижения"""
        meta = request.meta
        user_uuid = meta.user_uuid
        achievement_name = meta.name

        print(f"💾 Saving metadata for: {user_uuid}/{achievement_name}")

        # Добавляем timestamp если нет
        if not meta.created_at:
            meta.created_at = datetime.now().isoformat()

        # Сохраняем в памяти (в продакшене - в БД)
        if user_uuid not in self.achievements_meta:
            self.achievements_meta[user_uuid] = {}

        self.achievements_meta[user_uuid][achievement_name] = meta

        print(f"✅ Metadata saved for {achievement_name}")
        return common_pb2.Empty()

    async def AddAchievementMeta(self, request, context):
        """Асинхронно удалить достижение из S3 и метаданные"""
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        s3_key = f"achievements/{user_uuid}/{achievement_name}"

        print(f"🗑️ Deleting achievement: {s3_key}")

        try:
            s3_client = await self.ensure_s3_client()

            # Удаляем файл из S3
            await s3_client.delete_object(
                Bucket=self.bucket_name,
                Key=s3_key
            )

            # Удаляем метаданные
            if (user_uuid in self.achievements_meta and
                    achievement_name in self.achievements_meta[user_uuid]):
                del self.achievements_meta[user_uuid][achievement_name]

            print(f"✅ Achievement deleted: {s3_key}")
            return common_pb2.Empty()

        except Exception as e:
            print(f"❌ Error deleting achievement: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Delete error: {str(e)}")


async def serve():
    """Запуск асинхронного GRPC сервера"""
    server = aio.server()
    service_pb2_grpc.add_AchievementServiceServicer_to_server(
        GrpcService(), server
    )
    server.add_insecure_port('[::]:50051')

    await server.start()
    print("🚀 Async GRPC Server running on port 50051")

    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        await server.stop(5)


if __name__ == '__main__':
    asyncio.run(serve())
