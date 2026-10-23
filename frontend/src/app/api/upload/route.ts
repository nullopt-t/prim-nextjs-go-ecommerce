import { NextRequest, NextResponse } from "next/server";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";
import { awsCredentialsProvider } from "@vercel/functions/oidc";

const region =
  process.env.AWS_REGION ||
  process.env.MINIO_PUBLIC_AWS_REGION ||
  "eu-west-1";

const roleArn =
  process.env.AWS_ROLE_ARN ||
  process.env.MINIO_PUBLIC_AWS_ROLE_ARN ||
  "arn:aws:iam::183197503538:role/Vercel/access-s3-bole-compass-b09e5061c4f2";

const bucket =
  process.env.S3_BUCKET_NAME ||
  process.env.MINIO_PUBLIC_S3_BUCKET_NAME ||
  "s3-bole-compass-b09e5061c4f2";

// Configure S3 client using Vercel OIDC role assumption
const s3 = new S3Client({
  region,
  credentials: roleArn
    ? awsCredentialsProvider({
        roleArn,
      })
    : undefined,
});

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData();
    const file = formData.get("file") as File | null;
    const folder = (formData.get("folder") as string) || "catalog";

    if (!file) {
      return NextResponse.json(
        { error: "No file provided" },
        { status: 400 }
      );
    }

    const safeName = file.name.replace(/[^a-zA-Z0-9.-]/g, "_");
    const uniqueKey = `${folder}/${Date.now()}-${safeName}`;
    const buffer = Buffer.from(await file.arrayBuffer());

    await s3.send(
      new PutObjectCommand({
        Bucket: bucket,
        Key: uniqueKey,
        Body: buffer,
        ContentType: file.type || "application/octet-stream",
      })
    );

    const publicUrl = `https://${bucket}.s3.${region}.amazonaws.com/${uniqueKey}`;

    return NextResponse.json({
      success: true,
      url: publicUrl,
      publicUrl,
      key: uniqueKey,
      bucket,
    });
  } catch (error: any) {
    console.error("Vercel S3 upload error:", error);
    return NextResponse.json(
      { error: error?.message || "Failed to upload file to S3" },
      { status: 500 }
    );
  }
}
