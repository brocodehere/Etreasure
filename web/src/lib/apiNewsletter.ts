export interface NewsletterSubscriber {
  id: number;
  email: string;
  subscribed_at: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateNewsletterRequest {
  email: string;
}

export interface NewsletterResponse {
  message: string;
  email: string;
}

export async function subscribeToNewsletter(email: string): Promise<NewsletterResponse> {
  const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
  
  try {
    const response = await fetch(`${apiUrl}/api/public/newsletter/subscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to subscribe: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Newsletter subscription error:', error);
    throw error;
  }
}
