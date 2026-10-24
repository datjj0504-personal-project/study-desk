import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import {
  createNote,
  deleteNote,
  getNote,
  listNotes,
  updateNote,
  type Note,
} from "../services/notes";
import { toErrorMessage } from "../utils/errors";

const AUTOSAVE_DELAY_MS = 700;

export function NotesPage() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saveState, setSaveState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const savingRef = useRef(false);
  const selectedNote = notes.find((note) => note.id === selectedId);
  const isDirty =
    selectedId === null
      ? title.length > 0 || body.length > 0
      : selectedNote !== undefined &&
        (title !== selectedNote.title || body !== selectedNote.body);

  useEffect(() => {
    let active = true;
    listNotes()
      .then((loadedNotes) => {
        if (!active) return;
        setNotes(loadedNotes);
        if (loadedNotes.length > 0) {
          setSelectedId(loadedNotes[0].id);
          setTitle(loadedNotes[0].title);
          setBody(loadedNotes[0].body);
        }
      })
      .catch((loadError: unknown) => {
        if (active) setError(toErrorMessage(loadError));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  function selectNote(note: Note) {
    setSelectedId(note.id);
    setTitle(note.title);
    setBody(note.body);
    setError(null);
    setSaveState("idle");
    setStatusMessage(null);
    setConfirmDelete(false);
  }

  function startNewNote() {
    setSelectedId(null);
    setTitle("");
    setBody("");
    setError(null);
    setSaveState("idle");
    setStatusMessage(null);
    setConfirmDelete(false);
  }

  function discardChanges() {
    if (selectedNote) {
      setTitle(selectedNote.title);
      setBody(selectedNote.body);
    } else {
      setTitle("");
      setBody("");
    }
    setError(null);
    setSaveState("idle");
    setStatusMessage(null);
  }

  const saveDraft = useCallback(async () => {
    if (savingRef.current || loading || !isDirty) return;
    if (!title.trim()) {
      setError("A title is required before this note can be saved.");
      setSaveState("error");
      return;
    }

    savingRef.current = true;
    setSaving(true);
    setSaveState("saving");
    setError(null);
    setStatusMessage(null);
    try {
      if (selectedId === null) {
        const created = await createNote({ title, body });
        setNotes((currentNotes) => [created, ...currentNotes]);
        setSelectedId(created.id);
        setSaveState("saved");
        setStatusMessage("Note created.");
      } else {
        const current = selectedNote;
        if (!current) {
          setError("The selected note is no longer available. Refresh the list and try again.");
          setSaveState("error");
          return;
        }
        await updateNote({ ...current, title, body });
        const updated = await getNote(selectedId);
        setNotes((currentNotes) => [
          updated,
          ...currentNotes.filter((note) => note.id !== selectedId),
        ]);
        setSaveState("saved");
        setStatusMessage("Saved.");
      }
    } catch (saveError) {
      setError(toErrorMessage(saveError));
      setSaveState("error");
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }, [body, isDirty, loading, selectedId, selectedNote, title]);

  useEffect(() => {
    if (loading || !isDirty || !title.trim() || saving || saveState === "error") {
      return;
    }
    const timeout = window.setTimeout(() => {
      void saveDraft();
    }, AUTOSAVE_DELAY_MS);
    return () => window.clearTimeout(timeout);
  }, [isDirty, loading, saveDraft, saveState, saving, title]);

  function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void saveDraft();
  }

  async function handleDelete() {
    if (selectedId === null || isDirty) return;
    setSaving(true);
    setError(null);
    setStatusMessage(null);
    setSaveState("idle");
    try {
      await deleteNote(selectedId);
      const remainingNotes = notes.filter((note) => note.id !== selectedId);
      setNotes(remainingNotes);
      setConfirmDelete(false);
      if (remainingNotes.length > 0) {
        selectNote(remainingNotes[0]);
      } else {
        startNewNote();
      }
      setStatusMessage("Note deleted.");
    } catch (deleteError) {
      setError(toErrorMessage(deleteError));
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="flex h-full min-h-[32rem] flex-col">
      <header className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Notes</h1>
          <p className="mt-1 text-sm text-slate-400">Create and edit your study notes.</p>
        </div>
        <button
          type="button"
          onClick={startNewNote}
          disabled={isDirty || saving}
          className="rounded-md bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
        >
          New note
        </button>
      </header>

      {error && (
        <p role="alert" className="mb-4 rounded-md border border-red-500/40 bg-red-950/40 p-3 text-sm text-red-200">
          {error}
        </p>
      )}
      {saving ? (
        <p role="status" className="mb-4 text-sm text-slate-300">Saving…</p>
      ) : isDirty ? (
        <p role="status" className="mb-4 text-sm text-amber-300">
          {title.trim() ? "Unsaved changes" : "Add a title to save this note."}
        </p>
      ) : saveState === "saved" || statusMessage ? (
        <p role="status" className="mb-4 text-sm text-emerald-300">
          {statusMessage ?? "Saved."}
        </p>
      ) : null}

      <div className="grid min-h-0 flex-1 gap-5 lg:grid-cols-[16rem_minmax(0,1fr)]">
        <aside className="rounded-xl border border-edge bg-panel p-3">
          <h2 className="px-2 py-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
            Your notes
          </h2>
          {loading ? (
            <p className="px-2 py-4 text-sm text-slate-400">Loading notes…</p>
          ) : notes.length === 0 ? (
            <p className="px-2 py-4 text-sm text-slate-400">No saved notes yet.</p>
          ) : (
            <ul className="flex flex-col gap-1">
              {notes.map((note) => (
                <li key={note.id}>
                  <button
                    type="button"
                    aria-current={selectedId === note.id ? "true" : undefined}
                    onClick={() => selectNote(note)}
                    disabled={isDirty || saving}
                    className={`w-full rounded-md px-3 py-2 text-left ${
                      selectedId === note.id
                        ? "bg-accent/20 text-white"
                        : "text-slate-300 hover:bg-edge/60"
                    }`}
                  >
                    <span className="block truncate text-sm font-medium">
                      {note.title || "Untitled note"}
                    </span>
                    <span className="mt-1 block truncate text-xs text-slate-400">
                      {note.body || "No content"}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </aside>

        <form
          onSubmit={handleSave}
          className="flex min-h-[24rem] flex-col rounded-xl border border-edge bg-panel p-5"
        >
          <label htmlFor="note-title" className="mb-2 text-sm font-medium text-slate-300">
            Title
          </label>
          <input
            id="note-title"
            value={title}
            onChange={(event) => {
              setTitle(event.target.value);
              setSaveState("idle");
              setError(null);
              setStatusMessage(null);
            }}
            required
            maxLength={200}
            disabled={saving}
            placeholder="Note title"
            className="mb-4 rounded-md border border-edge bg-surface px-3 py-2 text-lg text-slate-100 outline-none focus:border-accent"
          />
          <label htmlFor="note-body" className="mb-2 text-sm font-medium text-slate-300">
            Body
          </label>
          <textarea
            id="note-body"
            value={body}
            onChange={(event) => {
              setBody(event.target.value);
              setSaveState("idle");
              setError(null);
              setStatusMessage(null);
            }}
            placeholder="Write your note…"
            disabled={saving}
            className="min-h-48 flex-1 resize-y rounded-md border border-edge bg-surface px-3 py-2 text-sm leading-6 text-slate-100 outline-none focus:border-accent"
          />
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <div className="flex flex-wrap items-center gap-2">
                {isDirty && (
                  <>
                    <span className="text-xs text-amber-300">Unsaved changes</span>
                    <button
                      type="button"
                      onClick={discardChanges}
                      disabled={saving || isDirty}
                      className="rounded-md px-3 py-2 text-sm text-slate-300 hover:bg-edge disabled:opacity-50"
                    >
                      Discard changes
                    </button>
                  </>
                )}
                {selectedId !== null &&
                  (confirmDelete ? (
                    <span className="flex items-center gap-2 text-sm">
                      <span>Delete this note?</span>
                      <button
                        type="button"
                        onClick={handleDelete}
                        disabled={saving || isDirty}
                        className="rounded-md bg-red-600 px-3 py-2 text-white hover:bg-red-500 disabled:opacity-50"
                      >
                        Confirm
                      </button>
                      <button
                        type="button"
                        onClick={() => setConfirmDelete(false)}
                        className="rounded-md px-3 py-2 text-slate-300 hover:bg-edge"
                      >
                        Cancel
                      </button>
                    </span>
                  ) : (
                    <button
                      type="button"
                      onClick={() => setConfirmDelete(true)}
                      disabled={saving}
                      className="rounded-md px-3 py-2 text-sm text-red-300 hover:bg-red-950/50 disabled:opacity-50"
                    >
                      Delete
                    </button>
                  ))}
              </div>
            </div>
            <button
              type="submit"
              disabled={saving || loading || !isDirty}
              className="rounded-md bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {saving ? "Saving…" : "Save now"}
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}
