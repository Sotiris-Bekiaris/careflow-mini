import { apiClient } from './api';
import type { Observation } from '@/stores/types';

const OBSERVATION_ENDPOINT = '/fhir/Observation';

/**
 * List all observations for a specific patient
 */
export const listPatientObservations = async (patientId: string): Promise<Observation[]> => {
  try {
    const response = await apiClient.get<{ entry: Array<{ resource: Observation }> }>(
      `${OBSERVATION_ENDPOINT}?patient=${patientId}`,
    );
    return response.data.entry?.map(entry => entry.resource) || [];
  } catch (error) {
    console.error(`Failed to list observations for patient ${patientId}:`, error);
    return [];
  }
};

/**
 * Get a specific observation by ID
 */
export const getObservation = async (id: string): Promise<Observation | null> => {
  try {
    const response = await apiClient.get<Observation>(`${OBSERVATION_ENDPOINT}/${id}`);
    return response.data;
  } catch (error) {
    console.error(`Failed to fetch observation ${id}:`, error);
    return null;
  }
};

/**
 * Create a new observation
 */
export const createObservation = async (
  observation: Partial<Observation>,
): Promise<Observation | null> => {
  try {
    const response = await apiClient.post<Observation>(OBSERVATION_ENDPOINT, observation);
    return response.data;
  } catch (error) {
    console.error('Failed to create observation:', error);
    return null;
  }
};
