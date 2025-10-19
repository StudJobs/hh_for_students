from aioitertools import asyncio

from achivment_service.app.minio.grpc_server import serve


if __name__ == '__main__':
    asyncio.run(serve())
