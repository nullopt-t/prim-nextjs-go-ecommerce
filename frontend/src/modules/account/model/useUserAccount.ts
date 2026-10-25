"use client";

import { useState, useEffect, useCallback } from "react";
import type { User } from "@/modules/auth/types";
import { userService } from "@/modules/account/api/userService";
import type {
  Address,
  CreateAddressPayload,
  PaymentMethod,
  UpdateAddressPayload,
} from "@/modules/account/types";
import type { Review } from "@/modules/catalog/types";

export function useAddresses() {
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAddresses = useCallback(async () => {
    try {
      const res = await userService.getAddresses();
      setAddresses(res?.data || []);
    } catch (err) {
      setError((err as { message?: string })?.message || "Failed to load addresses");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAddresses();
  }, [fetchAddresses]);

  const addAddress = async (payload: CreateAddressPayload) => {
    try {
      const res = await userService.createAddress(payload);
      await fetchAddresses();
      return { success: true, data: res?.data };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const updateAddress = async (id: string, payload: UpdateAddressPayload) => {
    try {
      const res = await userService.updateAddress(id, payload);
      await fetchAddresses();
      return { success: true, data: res?.data };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const setDefaultAddress = async (id: string) => {
    try {
      const res = await userService.setDefaultAddress(id);
      await fetchAddresses();
      return { success: true, data: res?.data };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const deleteAddress = async (id: string) => {
    try {
      await userService.deleteAddress(id);
      await fetchAddresses();
      return { success: true };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  return {
    addresses,
    loading,
    error,
    refreshAddresses: fetchAddresses,
    addAddress,
    updateAddress,
    setDefaultAddress,
    deleteAddress,
  };
}

export function usePaymentMethods() {
  const [cards, setCards] = useState<PaymentMethod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchCards = useCallback(async () => {
    try {
      const data = await userService.getPaymentMethods();
      setCards(data || []);
    } catch (err) {
      setError((err as { message?: string })?.message || "Failed to load payment methods");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCards();
  }, [fetchCards]);

  const addCard = async (payload: Record<string, unknown>) => {
    try {
      const newCard = await userService.createPaymentMethod(payload);
      await fetchCards();
      return { success: true, data: newCard };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const deleteCard = async (id: string) => {
    try {
      await userService.deletePaymentMethod(id);
      await fetchCards();
      return { success: true };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const setDefaultCard = async (id: string) => {
    try {
      await userService.setDefaultPaymentMethod(id);
      await fetchCards();
      return { success: true };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
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
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchReviews = useCallback(async () => {
    try {
      const res = await userService.getReviews();
      const list = Array.isArray(res) ? res : res?.data || [];
      setReviews(list);
    } catch (err) {
      setError((err as { message?: string })?.message || "Failed to load reviews");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReviews();
  }, [fetchReviews]);

  const toggleHelpful = async (id: string) => {
    try {
      const res = (await userService.markReviewHelpful(id)) as {
        likes?: number;
        userLiked?: boolean;
      } | undefined;
      setReviews((prev) =>
        prev.map((r) =>
          r.id === id ? ({ ...r, likes: res?.likes, userLiked: res?.userLiked } as Review) : r
        )
      );
      return { success: true };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
    }
  };

  const deleteReview = async (id: string) => {
    try {
      await userService.deleteReview(id);
      await fetchReviews();
      return { success: true };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
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
  const [profile, setProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = useCallback(async () => {
    try {
      const data = await userService.getProfile();
      setProfile(data?.data ?? null);
    } catch (err) {
      setError((err as { message?: string })?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  const updateProfile = async (payload: Record<string, unknown>) => {
    try {
      const updated = await userService.updateProfile(payload);
      setProfile(updated);
      return { success: true, data: updated };
    } catch (err) {
      return { success: false, error: (err as { message?: string })?.message };
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
