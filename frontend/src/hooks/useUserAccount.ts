"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/api/client";

export function useAddresses() {
  const [addresses, setAddresses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAddresses = useCallback(async () => {
    try {
      const data = await api.get("/api/v1/user/addresses");
      setAddresses(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load addresses");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    api
      .get("/api/v1/user/addresses")
      .then((data) => {
        if (isMounted) setAddresses(data || []);
      })
      .catch((err: any) => {
        if (isMounted) setError(err?.message || "Failed to load addresses");
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, []);

  const addAddress = async (payload: any) => {
    try {
      const newAddr = await api.post("/api/v1/user/addresses", payload);
      await fetchAddresses();
      return { success: true, data: newAddr };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const updateAddress = async (id: string, payload: any) => {
    try {
      const updated = await api.patch(`/api/v1/user/addresses/${id}`, payload);
      await fetchAddresses();
      return { success: true, data: updated };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteAddress = async (id: string) => {
    try {
      await api.delete(`/api/v1/user/addresses/${id}`);
      await fetchAddresses();
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  return {
    addresses,
    loading,
    error,
    refreshAddresses: fetchAddresses,
    addAddress,
    updateAddress,
    deleteAddress,
  };
}

export function usePaymentMethods() {
  const [cards, setCards] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchCards = useCallback(async () => {
    try {
      const data = await api.get("/api/v1/user/payment-methods");
      setCards(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load payment methods");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    api
      .get("/api/v1/user/payment-methods")
      .then((data) => {
        if (isMounted) setCards(data || []);
      })
      .catch((err: any) => {
        if (isMounted) setError(err?.message || "Failed to load payment methods");
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, []);

  const addCard = async (payload: any) => {
    try {
      const newCard = await api.post("/api/v1/user/payment-methods", payload);
      await fetchCards();
      return { success: true, data: newCard };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteCard = async (id: string) => {
    try {
      await api.delete(`/api/v1/user/payment-methods/${id}`);
      await fetchCards();
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const setDefaultCard = async (id: string) => {
    try {
      await api.patch(`/api/v1/user/payment-methods/${id}/default`);
      await fetchCards();
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  return {
    cards,
    loading,
    error,
    refreshCards: fetchCards,
    addCard,
    deleteCard,
    setDefaultCard,
  };
}

export function useUserReviews() {
  const [reviews, setReviews] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchReviews = useCallback(async () => {
    try {
      const data = await api.get("/api/v1/user/reviews");
      setReviews(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load reviews");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    api
      .get("/api/v1/user/reviews")
      .then((data) => {
        if (isMounted) setReviews(data || []);
      })
      .catch((err: any) => {
        if (isMounted) setError(err?.message || "Failed to load reviews");
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, []);

  const toggleHelpful = async (id: string) => {
    try {
      const res: any = await api.post(`/api/v1/user/reviews/${id}/helpful`);
      setReviews((prev) =>
        prev.map((r) => (r.id === id ? { ...r, likes: res.likes, userLiked: res.userLiked } : r))
      );
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteReview = async (id: string) => {
    try {
      await api.delete(`/api/v1/user/reviews/${id}`);
      await fetchReviews();
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  return {
    reviews,
    loading,
    error,
    refreshReviews: fetchReviews,
    toggleHelpful,
    deleteReview,
  };
}

export function useUserProfile() {
  const [profile, setProfile] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = useCallback(async () => {
    try {
      const data = await api.get("/api/v1/auth/me");
      setProfile(data);
    } catch (err: any) {
      setError(err?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    api
      .get("/api/v1/auth/me")
      .then((data) => {
        if (isMounted) setProfile(data);
      })
      .catch((err: any) => {
        if (isMounted) setError(err?.message || "Failed to load profile");
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, []);

  const updateProfile = async (payload: any) => {
    try {
      const updated = await api.patch("/api/v1/auth/me", payload);
      setProfile(updated);
      return { success: true, data: updated };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  return {
    profile,
    loading,
    error,
    refreshProfile: fetchProfile,
    updateProfile,
  };
}
