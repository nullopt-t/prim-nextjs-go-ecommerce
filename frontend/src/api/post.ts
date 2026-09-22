const getBaseUrl = () => {
  if (typeof window !== "undefined") {
    return process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000/api";
  }
  return process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000/api";
};

export async function postData<T = any, B = any>(
  endpoint: string,
  payload: B,
  options?: RequestInit
): Promise<T> {
  const url = `${getBaseUrl()}${endpoint}`;
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
      body: JSON.stringify(payload),
      ...options,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return (await response.json()) as T;
  } catch (err) {
    throw new Error(`Failed to POST ${endpoint}`, { cause: err });
  }
}

export default postData;
