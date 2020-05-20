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
        typedef struct {
            char *json;
            char *err;
        } parseResponse;
 
        parseResponse Parse(char* a);
        void free(void *ptr);
        """
)
ffi.compile()
