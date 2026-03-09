from cffi import FFI

ffi = FFI()

ffi.set_source(
    "pygohcl._pygohcl",
    None,
    include_dirs=[],
    extra_compile_args=["-march=native"],
    libraries=[],
)

ffi.cdef(
    """
        #if defined(_WIN32)
        #  define CFFI_DLLEXPORT  __declspec(dllexport)
        #elif defined(__GNUC__)
        #  define CFFI_DLLEXPORT  __attribute__((visibility("default")))
        #else
        #  define CFFI_DLLEXPORT  /* nothing */
        #endif
        
        typedef struct {
            char *json;
            char *err;
        } parseResponse;
 
        parseResponse Parse(char* a, int keepInterpFlag);
        parseResponse ParseAttributes(char* a);
        char* EvalValidationRule(char* c, char* e, char* n, char* v);
        CFFI_DLLEXPORT void free(void *ptr);
        """
)
ffi.compile()
