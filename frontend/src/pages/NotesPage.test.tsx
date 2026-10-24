import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NotesPage } from "./NotesPage";

const { api } = vi.hoisted(() => ({
  api: {
    createNote: vi.fn(),
    deleteNote: vi.fn(),
    listNotes: vi.fn(),
    updateNote: vi.fn(),
  },
}));

vi.mock("../services/notes", () => api);

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("NotesPage", () => {
  it("loads, creates, and selects a note", async () => {
    api.listNotes.mockResolvedValueOnce([]).mockResolvedValue([
      { id: 1, title: "Revision", body: "Chapter one", createdAt: "", updatedAt: "" },
    ]);
    api.createNote.mockResolvedValue({
      id: 1,
      title: "Revision",
      body: "Chapter one",
      createdAt: "",
      updatedAt: "",
    });

    render(<NotesPage />);
    expect(await screen.findByText("No saved notes yet.")).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Revision" } });
    fireEvent.change(screen.getByLabelText("Body"), { target: { value: "Chapter one" } });
    fireEvent.click(screen.getByRole("button", { name: "Create note" }));

    expect((await screen.findByRole("status")).textContent).toContain("Note created.");
    expect(api.createNote).toHaveBeenCalledWith({ title: "Revision", body: "Chapter one" });
    expect(screen.getByRole("button", { name: /Revision/ })).toBeTruthy();
  });

  it("updates and deletes a selected note", async () => {
    api.listNotes.mockResolvedValue([
      { id: 4, title: "Existing", body: "Old body", createdAt: "", updatedAt: "" },
    ]);
    api.updateNote.mockResolvedValue(undefined);
    api.deleteNote.mockResolvedValue(undefined);

    render(<NotesPage />);
    fireEvent.change(await screen.findByLabelText("Body"), {
      target: { value: "New body" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    await waitFor(() => expect(api.updateNote).toHaveBeenCalledWith(
      expect.objectContaining({ id: 4, title: "Existing", body: "New body" }),
    ));

    fireEvent.click(screen.getByRole("button", { name: "Delete" }));
    fireEvent.click(screen.getByRole("button", { name: "Confirm" }));
    await waitFor(() => expect(api.deleteNote).toHaveBeenCalledWith(4));
    expect(await screen.findByText("Note deleted.")).toBeTruthy();
  });

  it("surfaces loading errors", async () => {
    api.listNotes.mockRejectedValue(new Error("storage failed"));

    render(<NotesPage />);

    expect((await screen.findByRole("alert")).textContent).toContain("storage failed");
  });
});
