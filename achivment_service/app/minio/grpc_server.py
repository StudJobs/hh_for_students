import asyncio
import grpc
from grpc import aio
from datetime import datetime
from app.service_pb2_grpc import add_AchievementServiceServicer_to_server, AchievementServiceServicer
from app.service_pb2 import (
    GetAllAchievementsRequest, 
    GetAchievementRequest,
    GetAchievementUploadRequest,
    AddAchievementMetaRequest,
    DeleteAchievementRequest
)
from app.types_pb2 import AchievementList, AchievementUrl, AchievementMeta
from app.common_pb2 import Empty
from app.minio.s3 import AsyncS3BucketService
from app.config import settings


class GrpcService(AchievementServiceServicer):
    def __init__(self, s3_service: AsyncS3BucketService):
        self.s3_service = s3_service
        # Временное хранилище метаданных
        self.achievements_meta = {}

    async def GetAllAchievements(self, request, context):
        """Получить все достижения пользователя"""
        user_uuid = request.user_uuid
        print(f"📋 Get achievements for user: {user_uuid}")
        
        achievement_list = AchievementList()
        
        try:
            if user_uuid in self.achievements_meta:
                for achievement in self.achievements_meta[user_uuid].values():
                    achievement_list.achievements.append(achievement)
            
            print(f"✅ Found {len(achievement_list.achievements)} achievements")
            return achievement_list
            
        except Exception as e:
            print(f"❌ Error getting achievements: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Error: {str(e)}")

    async def GetAchievementDownloadUrl(self, request, context):
        """Получить URL для скачивания достижения из S3"""
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        
        print(f"📥 Generating download URL for: {user_uuid}/{achievement_name}")
        
        try:
            # ✅ РЕАЛЬНОЕ подключение к S3 через ваш сервис
            download_url = await self.s3_service.generate_presigned_url(
                prefix=f"achievements/{user_uuid}",
                source_file_name=achievement_name,
                expiration=3600,
                method="get_object"
            )
            
            achievement_url = AchievementUrl(
                url=download_url,
                method="GET",
                headers={"Content-Type": "application/octet-stream"}
            )
            
            print(f"✅ Download URL generated: {download_url[:80]}...")
            return achievement_url
            
        except Exception as e:
            print(f"❌ Error generating download URL: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"S3 error: {str(e)}")

    async def GetAchievementUploadUrl(self, request, context):
        """Получить URL для загрузки достижения в S3"""
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        file_name = request.file_name
        file_type = request.file_type
        
        print(f"📤 Generating upload URL for: {user_uuid}/{achievement_name}")
        
        try:
            # ✅ РЕАЛЬНОЕ подключение к S3 через ваш сервис
            upload_url = await self.s3_service.generate_presigned_url(
                prefix=f"achievements/{user_uuid}",
                source_file_name=achievement_name,
                expiration=3600,
                method="put_object"
            )
            
            achievement_url = AchievementUrl(
                url=upload_url,
                method="PUT",
                headers={"Content-Type": file_type}
            )
            
            print(f"✅ Upload URL generated: {upload_url[:80]}...")
            return achievement_url
            
        except Exception as e:
            print(f"❌ Error generating upload URL: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"S3 error: {str(e)}")

    async def AddAchievementMeta(self, request, context):
        """Добавить метаданные достижения"""
        try:
            meta = request.meta
            
            if not meta:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "Meta data is required")
                return Empty()
            
            if not hasattr(meta, 'user_uuid') or not meta.user_uuid:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "user_uuid is required")
                return Empty()
                
            if not hasattr(meta, 'name') or not meta.name:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "achievement name is required")
                return Empty()
            
            user_uuid = meta.user_uuid
            achievement_name = meta.name
            
            print(f"💾 Saving metadata for: {user_uuid}/{achievement_name}")
            
            # Добавляем timestamp если нет
            if not getattr(meta, 'created_at', None):
                meta.created_at = datetime.now().isoformat()
            
            # Сохраняем метаданные
            if user_uuid not in self.achievements_meta:
                self.achievements_meta[user_uuid] = {}
            
            self.achievements_meta[user_uuid][achievement_name] = meta
            
            print(f"✅ Metadata saved for {achievement_name}")
            return Empty()
            
        except Exception as e:
            print(f"❌ Error in AddAchievementMeta: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Error saving metadata: {str(e)}")

    async def DeleteAchievement(self, request, context):
        """Удалить достижение из S3"""
        user_uuid = request.user_uuid
        achievement_name = request.achievement_name
        
        print(f"🗑️ Deleting achievement: {user_uuid}/{achievement_name}")
        
        try:
            # ✅ РЕАЛЬНОЕ удаление из S3 через ваш сервис
            await self.s3_service.delete_file_object(
                prefix=f"achievements/{user_uuid}",
                source_file_name=achievement_name
            )
            
            # Удаляем метаданные
            if (user_uuid in self.achievements_meta and 
                achievement_name in self.achievements_meta[user_uuid]):
                del self.achievements_meta[user_uuid][achievement_name]
            
            print(f"✅ Achievement deleted from S3")
            return Empty()
            
        except Exception as e:
            print(f"❌ Error deleting achievement: {e}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Delete error: {str(e)}")


async def serve():
    """Запуск асинхронного GRPC сервера"""
    
    # ✅ РЕАЛЬНОЕ создание S3 сервиса с вашими настройками
    s3_service = AsyncS3BucketService(
        bucket_name=settings.s3.bucket_name,
        endpoint=settings.s3.endpoint,
        access_key=settings.s3.access_key,
        secret_key=settings.s3.secret_key,
    )
    
    # Создаем GRPC сервис и передаем реальный S3 сервис
    grpc_service = GrpcService(s3_service=s3_service)
    
    server = aio.server()
    add_AchievementServiceServicer_to_server(grpc_service, server)
    server.add_insecure_port('[::]:50053')
    
    await server.start()
    print("🚀 Async GRPC Server running on port 50053")
    
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        await server.stop(5)