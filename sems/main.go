package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"sems/alloc/pool"
	. "sems/http"
	. "sems/std"
)

func Page(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	default:
		switch path {
		case "/":
			return PageIndex(w, r)
		}
	case strings.HasPrefix(path, "/course"):
		switch path[len("/course"):] {
		default:
			return PageCourse(w, r)
		case "/create", "/edit":
			return PageCourseCreateEdit(w, r)
		case "/lesson":
			return PageCourseLesson(w, r)
		}
	case strings.HasPrefix(path, "/group"):
		switch path[len("/group"):] {
		default:
			return PageGroup(w, r)
		case "/create":
			return PageGroupCreate(w, r)
		case "/edit":
			return PageGroupEdit(w, r)
		}
	case strings.HasPrefix(path, "/subject"):
		path = path[len("/subject"):]

		switch {
		default:
			switch path {
			default:
				return PageSubject(w, r)
			case "/create":
				return PageSubjectCreate(w, r)
			case "/edit":
				return PageSubjectEdit(w, r)
			}
		case strings.HasPrefix(path, "/lesson"):
			switch path[len("/lesson"):] {
			default:
				return PageSubjectLesson(w, r)
			case "/edit":
				return PageSubjectLessonEdit(w, r)
			}
		}
	case strings.HasPrefix(path, "/submission"):
		switch path[len("/submission"):] {
		default:
			return PageSubmission(w, r)
		case "/new":
			return PageSubmissionNew(w, r)
		}
	case strings.HasPrefix(path, "/user"):
		switch path[len("/user"):] {
		default:
			return PageUser(w, r)
		case "/create":
			return PageUserCreate(w, r)
		case "/edit":
			return PageUserEdit(w, r)
		case "/signin":
			return PageUserSignin(w, r)
		}
	}
	return NotFound("requested page does not exist")
}

func Handler(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	case strings.HasPrefix(path, "/course"):
		switch path[len("/course"):] {
		case "/delete":
			return HandlerCourseDelete(w, r)
		}
	case strings.HasPrefix(path, "/group"):
		switch path[len("/group"):] {
		case "/create":
			return HandlerGroupCreate(w, r)
		case "/edit":
			return HandlerGroupEdit(w, r)
		}
	case strings.HasPrefix(path, "/subject"):
		switch path[len("/subject"):] {
		case "/create":
			return HandlerSubjectCreate(w, r)
		case "/edit":
			return HandlerSubjectEdit(w, r)
		}
	case strings.HasPrefix(path, "/user"):
		switch path[len("/user"):] {
		case "/create":
			return HandlerUserCreate(w, r)
		case "/edit":
			return HandlerUserEdit(w, r)
		case "/signin":
			return HandlerUserSignin(w, r)
		case "/signout":
			return HandlerUserSignout(w, r)
		}
	}
	return NotFound("requested API endpoint does not exist")
}

func router(w *HTTPResponse, r *HTTPRequest) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanicError(p)
		}
	}()

	path := r.URL.Path
	switch {
	default:
		return Page(w, r, path)
	case strings.HasPrefix(path, "/api"):
		return Handler(w, r, path[len("/api"):])
	case path == "/error":
		return ServerError(NewError("test error"))
	case path == "/panic":
		panic("test panic")
	}
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	if r.URL.Path == "/hello" {
		w.AppendString("Hello, world!\n")
		return
	}

	level := "[DEBUG]"
	start := time.Now()

	err := router(w, r)
	if err != nil {
		var message string
		if httpError := (HTTPError{}); errors.As(err, &httpError) {
			w.StatusCode = httpError.StatusCode
			message = httpError.DisplayMessage
			if w.StatusCode >= HTTPStatusBadRequest && w.StatusCode < HTTPStatusInternalServerError {
				// 4xx
				level = "[WARN]"
			} else {
				level = "[ERROR]"
			}
		} else if errors.As(err, new(PanicError)) {
			w.StatusCode = ServerError(nil).StatusCode
			message = ServerError(nil).DisplayMessage
			level = "[ERROR]"
		} else {
			log.Panicf("Unsupported error type %T", err)
		}

		ErrorPageHandler(w, message)
	}

	log.Println(level, "[%21s] %7s %s -> %d (%v), %v", r.Address, r.Method, r.URL.Path, w.StatusCode, err, time.Since(start))
}

func run() error {
	log.Println("[INFO] Starting SEMS")

	if err := RestoreSessionsFromFile(SessionsFile); err != nil {
		log.Println("[WARN] Failed to restore sessions from file:", err.Error())
	}
	if err := RestoreDBFromFile(DBFile); err != nil {
		log.Println("[WARN] Failed to restore DB from file:", err.Error())
		CreateInitialDB()
	}

	l, err := TCPListen(7072)
	if err != nil {
		return fmt.Errorf("Failed to listen on port: %w", err)
	}
	log.Println("[INFO] Listening on 0.0.0.0:7072...")

	q, err := NewEventQueue()
	if err != nil {
		return fmt.Errorf("Failed to create event queue: %w", err)
	}

	q.AddSocket(l, EventRequestRead)

	signal.Ignore(syscall.Signal(SIGINT), syscall.Signal(SIGTERM))
	if err := q.AddSignal(SIGINT, SIGTERM); err != nil {
		return fmt.Errorf("subscribe to signals: %w", err)
	}

	ctxPool := pool.New(NewHTTPContext, (*HTTPContext).Reset)

	var quit bool
	for !quit {
		event, err := q.GetEvent()
		if err != nil {
			log.Println("[ERROR] Failed to get event:", err.Error())
			continue
		}

		switch event.Type {
		default:
			log.Println("[ERROR] Unhandled event", event.Type)
		case EventRead:
			switch event.Fd {
			case l: /* ready to accept new connection. */
				ctx, c, err := HTTPAccept(l, ctxPool)
				if err != nil {
					log.Println("[ERROR] Failed to accept new HTTP connection:", err.Error())
					continue
				}

				var tp Timespec
				if err := ClockGettime(CLOCK_REALTIME, &tp); err != nil {
					return fmt.Errorf("Failed to get current walltime: %w", err)
				}
				tp.Nsec = 0 /* NOTE: we don't care about nanoseconds. */
				dateBuf := unsafe.Slice(&ctx.DateBuf[0], len(ctx.DateBuf))
				SlicePutTmRFC822(dateBuf, TimeToTm(int(tp.Sec)))

				q.AddSocket(c, EventRequestRead|EventRequestWrite)
			default: /* ready to serve new HTTP request. */
				ctx, check := HTTPContextFromCheckedPointer(event.UserData)
				if ctx.Check != check {
					continue
				}

				if event.EndOfFile {
					ctxPool.Put(ctx)
					Close(event.Fd)
					continue
				}

				if err := HTTPRead(ctx, event.Fd); err != nil {
					log.Println("[ERROR] Failed to read data from socket:", err.Error())
					ctxPool.Put(ctx)
					Close(event.Fd)
					continue
				}

				actx := StartAccept(ctx)
				for {
					w, r, ok := actx.Read()
					if !ok {
						break
					}

					Router(w, r)
					actx.Done()
				}

				if err := HTTPWrite(ctx, event.Fd); err != nil {
					log.Println("[ERROR] Failed to write HTTP response:", err.Error())
					ctxPool.Put(ctx)
					Close(event.Fd)
				}
			}
		case EventWrite:
			ctx, check := HTTPContextFromCheckedPointer(event.UserData)
			if ctx.Check != check {
				continue
			}

			if event.EndOfFile {
				ctxPool.Put(ctx)
				Close(event.Fd)
				continue
			}

			if err := HTTPWrite(ctx, event.Fd); err != nil {
				log.Println("[ERROR] Failed to write HTTP response:", err.Error())
				ctxPool.Put(ctx)
				Close(event.Fd)
			}
		case EventSignal:
			log.Printf("[INFO] Received signal %d, exitting...\n", event.Fd)
			quit = true
		}
	}

	q.Close()
	Close(l)

	if err := StoreDBToFile(DBFile); err != nil {
		log.Println("[WARN] Failed to store DB to file:", err.Error())
	}
	if err := StoreSessionsToFile(SessionsFile); err != nil {
		log.Println("[WARN] Failed to store sessions to file:", err.Error())
	}

	return nil
}

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
