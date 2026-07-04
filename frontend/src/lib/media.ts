// lib/media.ts
// Direct-to-storage image upload via presigned URLs (backend: media module).
//
// Flow: ask the backend for a presigned PUT URL, upload the file bytes
// straight to MinIO (not through our server), then use the returned public
// URL as the image reference. See docs/architecture.md Media module.

import { authPost } from "./api";

interface UploadURLResponse {
  upload_url: string;
  public_url: string;
}

/**
 * Uploads an image file and resolves to its public URL.
 * Throws on non-image types or upload failure.
 */
export async function uploadImage(file: File): Promise<string> {
  const { upload_url, public_url } = await authPost<UploadURLResponse>("/media/upload-url", {
    content_type: file.type,
  });

  const put = await fetch(upload_url, {
    method: "PUT",
    headers: { "Content-Type": file.type },
    body: file,
  });
  if (!put.ok) {
    throw new Error(`Upload failed (${put.status})`);
  }

  return public_url;
}
