"use client";

import { useState, useEffect, useCallback } from "react";
import { userService } from "@/services/user";

export function useAddresses() {
  const [addresses, setAddresses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAddresses = useCallback(async () => {
    try {
      const data = await userService.getAddresses();
      setAddresses(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load addresses");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAddresses();
  }, [fetchAddresses]);

  const addAddress = async (payload: any) => {
    try {
      const newAddr = await userService.createAddress(payload);
      await fetchAddresses();
      return { success: true, data: newAddr };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const updateAddress = async (id: string, payload: any) => {
    try {
      const updated = await userService.updateAddress(id, payload);
      await fetchAddresses();
      return { success: true, data: updated };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteAddress = async (id: string) => {
    try {
      await userService.deleteAddress(id);
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
      const data = await userService.getPaymentMethods();
      setCards(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load payment methods");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCards();
  }, [fetchCards]);

  const addCard = async (payload: any) => {
    try {
      const newCard = await userService.createPaymentMethod(payload);
      await fetchCards();
      return { success: true, data: newCard };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteCard = async (id: string) => {
    try {
      await userService.deletePaymentMethod(id);
      await fetchCards();
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const setDefaultCard = async (id: string) => {
    try {
      await userService.setDefaultPaymentMethod(id);
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
      const data = await userService.getReviews();
      setReviews(data || []);
    } catch (err: any) {
      setError(err?.message || "Failed to load reviews");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReviews();
  }, [fetchReviews]);

  const toggleHelpful = async (id: string) => {
    try {
      const res: any = await userService.markReviewHelpful(id);
      setReviews((prev) =>
        prev.map((r) => (r.id === id ? { ...r, likes: res?.likes, userLiked: res?.userLiked } : r))
      );
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err?.message };
    }
  };

  const deleteReview = async (id: string) => {
    try {
      await userService.deleteReview(id);
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
      const data = await userService.getProfile();
      setProfile(data);
    } catch (err: any) {
      setError(err?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  const updateProfile = async (payload: any) => {
    try {
      const updated = await userService.updateProfile(payload);
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
