// SPDX-License-Identifier: Unlicense OR MIT

#include <jni.h>
#include "os_android.h"
#include "_cgo_export.h"

jint gio_jni_GetEnv(JavaVM *vm, JNIEnv **env, jint version) {
	return (*vm)->GetEnv(vm, (void **)env, version);
}

jint gio_jni_GetJavaVM(JNIEnv *env, JavaVM **jvm) {
	return (*env)->GetJavaVM(env, jvm);
}

jint gio_jni_AttachCurrentThread(JavaVM *vm, JNIEnv **p_env, void *thr_args) {
	return (*vm)->AttachCurrentThread(vm, p_env, thr_args);
}

jint gio_jni_DetachCurrentThread(JavaVM *vm) {
	return (*vm)->DetachCurrentThread(vm);
}

jobject gio_jni_NewGlobalRef(JNIEnv *env, jobject obj) {
	return (*env)->NewGlobalRef(env, obj);
}

void gio_jni_DeleteGlobalRef(JNIEnv *env, jobject obj) {
	(*env)->DeleteGlobalRef(env, obj);
}

jclass gio_jni_GetObjectClass(JNIEnv *env, jobject obj) {
	return (*env)->GetObjectClass(env, obj);
}

jmethodID gio_jni_GetMethodID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetMethodID(env, clazz, name, sig);
}

jmethodID gio_jni_GetStaticMethodID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetStaticMethodID(env, clazz, name, sig);
}

jfloat gio_jni_CallFloatMethod(JNIEnv *env, jobject obj, jmethodID methodID) {
	return (*env)->CallFloatMethod(env, obj, methodID);
}

jint gio_jni_CallIntMethod(JNIEnv *env, jobject obj, jmethodID methodID) {
	return (*env)->CallIntMethod(env, obj, methodID);
}

void gio_jni_CallVoidMethod(JNIEnv *env, jobject obj, jmethodID methodID, const jvalue *args) {
	(*env)->CallVoidMethodA(env, obj, methodID, args);
}

jbyte *gio_jni_GetByteArrayElements(JNIEnv *env, jbyteArray arr) {
	return (*env)->GetByteArrayElements(env, arr, NULL);
}

void gio_jni_ReleaseByteArrayElements(JNIEnv *env, jbyteArray arr, jbyte *bytes) {
	(*env)->ReleaseByteArrayElements(env, arr, bytes, JNI_ABORT);
}

jsize gio_jni_GetArrayLength(JNIEnv *env, jbyteArray arr) {
	return (*env)->GetArrayLength(env, arr);
}

jstring gio_jni_NewString(JNIEnv *env, const jchar *unicodeChars, jsize len) {
	return (*env)->NewString(env, unicodeChars, len);
}
