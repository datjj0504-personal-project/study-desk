import {
  CreateNote,
  DeleteNote,
  GetNote,
  ListNotes,
  UpdateNote,
} from "../../wailsjs/go/main/App";
import { notes as generatedNotes } from "../../wailsjs/go/models";

export interface Note {
  id: number;
  title: string;
  body: string;
  subjectId?: number;
  createdAt: string;
  updatedAt: string;
}

export interface NoteInput {
  title: string;
  body: string;
  subjectId?: number;
}

function toNote(note: generatedNotes.Note): Note {
  return {
    id: note.id,
    title: note.title,
    body: note.body,
    subjectId: note.subjectId,
    createdAt: String(note.createdAt),
    updatedAt: String(note.updatedAt),
  };
}

export async function listNotes(): Promise<Note[]> {
  return (await ListNotes()).map(toNote);
}

export async function getNote(id: number): Promise<Note> {
  return toNote(await GetNote(id));
}

export async function createNote(input: NoteInput): Promise<Note> {
  return toNote(
    await CreateNote(new generatedNotes.Note({
      title: input.title,
      body: input.body,
      subjectId: input.subjectId,
    })),
  );
}

export async function updateNote(note: Note): Promise<void> {
  await UpdateNote(new generatedNotes.Note(note));
}

export async function deleteNote(id: number): Promise<void> {
  await DeleteNote(id);
}
