import asyncio
import grpc
from grpc import aio
import aiohttp
import os
from pathlib import Path

from app.service_pb2 import (
    GetAllAchievementsRequest,
    GetAchievementRequest,
    GetAchievementUploadRequest, 
    AddAchievementMetaRequest,
    DeleteAchievementRequest
)
from app.types_pb2 import AchievementMeta
from app.service_pb2_grpc import AchievementServiceStub

async def upload_real_file(upload_url: str, file_path: str, content_type: str):
    """Загружает реальный файл по presigned URL"""
    try:
        async with aiohttp.ClientSession() as session:
            with open(file_path, 'rb') as file:
                file_data = file.read()
                
            headers = {
                'Content-Type': content_type,
            }
            
            print(f"📤 Uploading {os.path.getsize(file_path)} bytes to S3...")
            async with session.put(upload_url, data=file_data, headers=headers) as response:
                if response.status == 200:
                    print(f"✅ File uploaded successfully to S3")
                    return True
                else:
                    print(f"❌ File upload failed: {response.status} - {await response.text()}")
                    return False
                    
    except Exception as e:
        print(f"❌ Error uploading file: {e}")
        return False

async def download_and_verify_file(download_url: str, original_path: str):
    """Скачивает файл из S3 и проверяет его целостность"""
    try:
        async with aiohttp.ClientSession() as session:
            print(f"📥 Downloading file from S3...")
            async with session.get(download_url) as response:
                if response.status == 200:
                    downloaded_content = await response.read()
                    
                    # Читаем оригинальный файл для сравнения
                    with open(original_path, 'rb') as f:
                        original_content = f.read()
                    
                    # Проверяем что файлы идентичны
                    if downloaded_content == original_content:
                        print(f"✅ File integrity verified: {len(downloaded_content)} bytes match original")
                        return True
                    else:
                        print(f"❌ File integrity check failed: sizes {len(downloaded_content)} vs {len(original_content)}")
                        return False
                else:
                    print(f"❌ Download failed: {response.status}")
                    return False
                    
    except Exception as e:
        print(f"❌ Error downloading file: {e}")
        return False

async def test_with_real_pdf_file():
    """Тест с реальным PDF файлом"""
    print("🚀 Testing with REAL PDF file...")
    
    # Создаем тестовый PDF файл если его нет
    test_file_path = "real_test_certificate.pdf"
    if not os.path.exists(test_file_path):
        # Создаем простой PDF
        pdf_content = b"%PDF-1.4\n1 0 obj\n<<>>\nendobj\n"
        with open(test_file_path, 'wb') as f:
            f.write(pdf_content)
        print(f"📄 Created test PDF file: {test_file_path}")
    
    file_size = os.path.getsize(test_file_path)
    print(f"📊 File size: {file_size} bytes")
    
    async with aio.insecure_channel('localhost:50051') as channel:
        stub = AchievementServiceStub(channel)
        
        user_uuid = "ed475a9e-7b59-4f51-a842-6da0ab33f79f"
        achievement_name = "Python_Developer_Certificate"
        
        try:
            # 1. Получаем URL для загрузки
            print("\n1. 📤 Getting upload URL for PDF...")
            upload_response = await stub.GetAchievementUploadUrl(GetAchievementUploadRequest(
                user_uuid=user_uuid,
                achievement_name=achievement_name,
                file_name="python_certificate.pdf",
                file_type="application/pdf"
            ))
            print(f"   ✅ Upload URL received")
            
            # 2. Загружаем реальный PDF файл в S3
            print("\n2. ⬆️ Uploading real PDF to S3...")
            upload_success = await upload_real_file(
                upload_response.url, 
                test_file_path,
                "application/pdf"
            )
            
            if not upload_success:
                print("❌ File upload failed, stopping test")
                return
            
            # 3. Сохраняем метаданные с реальными данными файла
            print("\n3. 💾 Saving metadata with real file info...")
            meta = AchievementMeta(
                name=achievement_name,
                user_uuid=user_uuid,
                file_name="python_certificate.pdf",
                file_type="application/pdf", 
                file_size=file_size,
                created_at="2024-01-15T10:30:00Z"  # реальная дата
            )
            await stub.AddAchievementMeta(AddAchievementMetaRequest(meta=meta))
            print("   ✅ Real metadata saved")
            
            # 4. Проверяем список достижений
            print("\n4. 📋 Checking achievements list...")
            list_response = await stub.GetAllAchievements(GetAllAchievementsRequest(user_uuid=user_uuid))
            print(f"   ✅ Found {len(list_response.achievements)} achievements")
            for achievement in list_response.achievements:
                print(f"      - {achievement.name} ({achievement.file_name}, {achievement.file_size} bytes)")
            
            # 5. Получаем URL для скачивания
            print("\n5. 📥 Getting download URL...")
            download_response = await stub.GetAchievementDownloadUrl(GetAchievementRequest(
                user_uuid=user_uuid,
                achievement_name=achievement_name
            ))
            print(f"   ✅ Download URL received")
            
            # 6. Скачиваем и проверяем файл из S3
            print("\n6. 🔍 Downloading and verifying file from S3...")
            download_success = await download_and_verify_file(download_response.url, test_file_path)
            
            if download_success:
                print("\n🎉 SUCCESS: Real file operations working perfectly!")
                print(f"   • File uploaded to S3: {achievement_name}")
                print(f"   • File downloaded from S3 and verified")
                print(f"   • Metadata stored correctly")
            else:
                print("\n❌ File verification failed")
            
            # 7. Очистка (опционально)
            print("\n7. 🗑️ Cleaning up...")
            # Удаляем тестовый файл
            if os.path.exists(test_file_path):
                os.remove(test_file_path)
                print("   ✅ Test file cleaned up")
            
            # Удаляем из S3 (раскомментируйте если нужно)
            # await stub.DeleteAchievement(DeleteAchievementRequest(
            #     user_uuid=user_uuid,
            #     achievement_name=achievement_name
            # ))
            # print("   ✅ Achievement deleted from S3")
            
        except Exception as e:
            print(f"❌ Test failed: {e}")
            import traceback
            traceback.print_exc()

async def test_with_image_file():
    """Тест с реальным изображением"""
    print("\n" + "="*50)
    print("🖼️ Testing with REAL image file...")
    
    # Создаем простой PNG файл
    test_image_path = "test_achievement.png"
    
    # Простой PNG заголовок (валидный PNG)
    png_header = b'\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x00\x00\x00\x00IEND\xaeB`\x82'
    
    with open(test_image_path, 'wb') as f:
        f.write(png_header)
    
    file_size = os.path.getsize(test_image_path)
    print(f"📊 Image size: {file_size} bytes")
    
    async with aio.insecure_channel('localhost:50051') as channel:
        stub = AchievementServiceStub(channel)
        
        user_uuid = "ed475a9e-7b59-4f51-a842-6da0ab33f79f"
        achievement_name = "Achievement_Badge"
        
        try:
            # Получаем URL для загрузки изображения
            upload_response = await stub.GetAchievementUploadUrl(GetAchievementUploadRequest(
                user_uuid=user_uuid,
                achievement_name=achievement_name,
                file_name="achievement_badge.png",
                file_type="image/png"
            ))
            
            # Загружаем изображение
            upload_success = await upload_real_file(upload_response.url, test_image_path, "image/png")
            
            if upload_success:
                # Сохраняем метаданные
                meta = AchievementMeta(
                    name=achievement_name,
                    user_uuid=user_uuid,
                    file_name="achievement_badge.png",
                    file_type="image/png", 
                    file_size=file_size
                )
                await stub.AddAchievementMeta(AddAchievementMetaRequest(meta=meta))
                print("✅ Image metadata saved")
                
                # Проверяем скачивание
                download_response = await stub.GetAchievementDownloadUrl(GetAchievementRequest(
                    user_uuid=user_uuid,
                    achievement_name=achievement_name
                ))
                
                download_success = await download_and_verify_file(download_response.url, test_image_path)
                if download_success:
                    print("🎉 Image test successful!")
            
            # Очистка
            if os.path.exists(test_image_path):
                os.remove(test_image_path)
                
        except Exception as e:
            print(f"❌ Image test failed: {e}")

if __name__ == '__main__':
    # Установите aiohttp если нужно
    # pip install aiohttp
    
    asyncio.run(test_with_real_pdf_file())
    asyncio.run(test_with_image_file())